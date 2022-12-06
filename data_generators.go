package mybench

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/atomic"
)

const characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// TODO: have the generators track the existing values by modifying this
// interface so existing values can be added and removed and SampleFromExisting
// only ever sample from known existing values?
type DataGenerator interface {
	Generate(*rand.Rand) interface{}
	SampleFromExisting(*rand.Rand) interface{}
}

// Always generates a null value. Sampling from it always returns a null value
// as well.
type NullGenerator struct {
}

func NewNullGenerator() *NullGenerator {
	return &NullGenerator{}
}

func (*NullGenerator) Generate(*rand.Rand) interface{} {
	return nil
}

func (*NullGenerator) SampleFromExisting(*rand.Rand) interface{} {
	return nil
}

// Generates an integer value in the inclusive range between min and max
// uniformly.
//
// Sampling from existing is the same as the generation, which mean it is not
// guaranteed to generate an existing value if the number of rows in the
// database is small.
type UniformIntGenerator struct {
	min int64
	max int64
}

func NewUniformIntGenerator(min, max int64) *UniformIntGenerator {
	return &UniformIntGenerator{min, max}
}

func (g *UniformIntGenerator) Generate(r *rand.Rand) interface{} {
	return g.GenerateTyped(r)
}

func (g *UniformIntGenerator) SampleFromExisting(r *rand.Rand) interface{} {
	return g.SampleFromExistingTyped(r)
}

func (g *UniformIntGenerator) GenerateTyped(r *rand.Rand) int64 {
	return r.Int63n(g.max-g.min+1) + g.min
}

func (g *UniformIntGenerator) SampleFromExistingTyped(r *rand.Rand) int64 {
	return g.GenerateTyped(r)
}

// Generates an random decimal value
//
// Sampling from existing is the same as the generation, which mean it is not
// guaranteed to generate an existing value if the number of rows in the
// database is small or the decimal has a large precision
type UniformDecimalGenerator struct {
	precision int
	scale     int
}

func NewUniformDecimalGenerator(precision, scale int) *UniformDecimalGenerator {
	return &UniformDecimalGenerator{precision, scale}
}

func (g *UniformDecimalGenerator) Generate(r *rand.Rand) interface{} {
	return g.GenerateTyped(r)
}

func (g *UniformDecimalGenerator) SampleFromExisting(r *rand.Rand) interface{} {
	return g.SampleFromExistingTyped(r)
}

func (g *UniformDecimalGenerator) GenerateTyped(r *rand.Rand) string {
	num := rand.Float64() * math.Pow10(g.precision) / math.Pow10(g.scale)
	format := fmt.Sprintf("%%%d.%df", g.precision, g.scale)
	return fmt.Sprintf(format, num)
}

func (g *UniformDecimalGenerator) SampleFromExistingTyped(r *rand.Rand) string {
	return g.GenerateTyped(r)
}

// Generates a floating point number with a given normal distribution.
//
// Sample from existing is the same as generating a number, which means it is
// not guaranteed to land on an existing value.
type NormalFloatGenerator struct {
	mean   float64
	stddev float64
}

func NewNormalFloatGenerator(mean, stddev float64) *NormalFloatGenerator {
	return &NormalFloatGenerator{mean, stddev}
}

func (g *NormalFloatGenerator) Generate(r *rand.Rand) interface{} {
	return g.GenerateTyped(r)
}

func (g *NormalFloatGenerator) SampleFromExisting(r *rand.Rand) interface{} {
	return g.SampleFromExistingTyped(r)
}

func (g *NormalFloatGenerator) GenerateTyped(r *rand.Rand) float64 {
	return r.NormFloat64()*g.stddev + g.mean
}

func (g *NormalFloatGenerator) SampleFromExistingTyped(r *rand.Rand) float64 {
	return g.GenerateTyped(r)
}

// Generates a string with a known amount of unique values (cardinality) and a
// fixed string length.
//
// The algorithm is fairly straight-forward: simply generate a number, convert
// it into a string, then repeat the letters until the desired length. Some
// hacky measures are taken such that multiple values (like 2 and 22) do not
// generate the same string, but there's no guarantee.
//
// Sample from existing is the same as generation, which means it is not
// guaranteed to land on an existing value.
type BoundedCardinalityStringGenerator struct {
	cardinality int
	length      int
}

func NewBoundedCardinalityStringGenerator(cardinality, length int) *BoundedCardinalityStringGenerator {
	return &BoundedCardinalityStringGenerator{cardinality, length}
}

func (g *BoundedCardinalityStringGenerator) Generate(r *rand.Rand) interface{} {
	return g.GenerateTyped(r)
}

func (g *BoundedCardinalityStringGenerator) SampleFromExisting(r *rand.Rand) interface{} {
	return g.SampleFromExistingTyped(r)
}

func (g *BoundedCardinalityStringGenerator) GenerateTyped(r *rand.Rand) string {
	n := r.Int63n(int64(g.cardinality))
	return generateUniqueStringFromInt(n, g.length)
}

func (g *BoundedCardinalityStringGenerator) SampleFromExistingTyped(r *rand.Rand) string {
	return g.GenerateTyped(r)
}

// Generates a string uniformly distributed between two lengths in a completely
// random fashion.
//
// Sample from existing is the same as generation and has a negligible chance of
// hitting an existing value.
type TotallyRandomStringGenerator struct {
	minLength int
	maxLength int
}

func NewTotallyRandomStringGenerator(minLength, maxLength int) *TotallyRandomStringGenerator {
	return &TotallyRandomStringGenerator{minLength, maxLength}
}

func (g *TotallyRandomStringGenerator) Generate(r *rand.Rand) interface{} {
	return g.GenerateTyped(r)
}

func (g *TotallyRandomStringGenerator) SampleFromExisting(r *rand.Rand) interface{} {
	return g.SampleFromExistingTyped(r)
}

func (g *TotallyRandomStringGenerator) GenerateTyped(r *rand.Rand) string {
	n := r.Intn(g.maxLength-g.minLength+1) + g.minLength

	buf := make([]byte, n)
	for i := range buf {
		buf[i] = characters[r.Intn(len(characters))]
	}

	return string(buf)
}

func (g *TotallyRandomStringGenerator) SampleFromExistingTyped(r *rand.Rand) string {
	return g.GenerateTyped(r)
}

// Generates UUIDs
// SampleFromExisting is basically broken as this should only very rarely
// generate a duplicate UUID.
// Version 1 uuid's have the timestamp at which they were generated embedded in them
// Version 4 uuid's are random
type UuidGenerator struct {
	Version int
}

// NewUuidGenerator
// Only version 1 (timebased) and version 4 (random) supported
func NewUuidGenerator(version int) *UuidGenerator {
	return &UuidGenerator{Version: version}
}

func (g *UuidGenerator) Generate(r *rand.Rand) interface{} {
	return g.GenerateTyped(r)
}

func (g *UuidGenerator) SampleFromExisting(r *rand.Rand) interface{} {
	return g.SampleFromExistingTyped(r)
}

func (g *UuidGenerator) GenerateTyped(r *rand.Rand) string {
	var u uuid.UUID
	if g.Version == 1 {
		u = uuid.Must(uuid.NewUUID())
	} else if g.Version == 4 {
		u = uuid.New()
	} else {
		panic("Only Supports type 1 or 4 UUIDs")
	}
	return u.String()
}

func (g *UuidGenerator) SampleFromExistingTyped(r *rand.Rand) string {
	return g.GenerateTyped(r)
}

// Atomically generate an auto incrementing value from the client-side.
//
// Sample from existing with sample uniformly between the min value to the
// current value. There is no guarantee that it will land on an existing value
// if values have been deleted.
//
// TODO: track deletion, but this is problematic too, because golang doesn't
// offer a concurrent-write map.
type AutoIncrementGenerator struct {
	min     int64
	current *atomic.Int64
}

func NewAutoIncrementGenerator(min, current int64) *AutoIncrementGenerator {
	return &AutoIncrementGenerator{min, atomic.NewInt64(current)}
}

func NewAutoIncrementGeneratorFromDatabase(databaseConfig DatabaseConfig, table, column string) (*AutoIncrementGenerator, error) {
	conn, err := databaseConfig.Connection()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	query := fmt.Sprintf("SELECT MIN(%s), MAX(%s) FROM %s.%s", column, column, databaseConfig.Database, table)
	res, err := conn.Execute(query)
	if err != nil {
		return nil, err
	}

	min, err := res.GetInt(0, 0)
	if err != nil {
		return nil, err
	}

	current, err := res.GetInt(0, 1)
	if err != nil {
		return nil, err
	}

	return &AutoIncrementGenerator{min, atomic.NewInt64(current)}, nil
}

func (g *AutoIncrementGenerator) Generate(r *rand.Rand) interface{} {
	return g.GenerateTyped(r)
}

func (g *AutoIncrementGenerator) SampleFromExisting(r *rand.Rand) interface{} {
	return g.SampleFromExistingTyped(r)
}

func (g *AutoIncrementGenerator) GenerateTyped(r *rand.Rand) int64 {
	return g.current.Add(1)
}

func (g *AutoIncrementGenerator) SampleFromExistingTyped(r *rand.Rand) int64 {
	return r.Int63n(g.Current()-g.min+1) + g.min
}

// Get the current value without generating a new value.
func (g *AutoIncrementGenerator) Current() int64 {
	return g.current.Load()
}

func (g *AutoIncrementGenerator) Min() int64 {
	return g.min
}

// Generates an unique string of some length with an incrementing counter
type AutoIncrementStringGenerator struct {
	min     int64
	current *atomic.Int64
	length  int
}

func NewAutoIncrementStringGenerator(min, current int64, length int) *AutoIncrementStringGenerator {
	return &AutoIncrementStringGenerator{
		min:     min,
		current: atomic.NewInt64(current),
		length:  length,
	}
}

func NewAutoIncrementStringGeneratorFromDatabase(conn *Connection, database, table, column string, length int) (*AutoIncrementStringGenerator, error) {
	query := fmt.Sprintf("SELECT MIN(CAST(SUBSTRING_INDEX(%s, '!', 1) AS UNSIGNED)) AS min_value, MAX(CAST(SUBSTRING_INDEX(%s, '!', 1) AS UNSIGNED)) AS current_value FROM %s.%s", column, column, database, table)
	res, err := conn.Execute(query)
	if err != nil {
		return nil, err
	}

	min, err := res.GetInt(0, 0)
	if err != nil {
		return nil, err
	}

	current, err := res.GetInt(0, 1)
	if err != nil {
		return nil, err
	}

	return &AutoIncrementStringGenerator{
		min:     min,
		current: atomic.NewInt64(current),
		length:  length,
	}, nil
}

func (g *AutoIncrementStringGenerator) Generate(r *rand.Rand) interface{} {
	return g.GenerateTyped(r)
}

func (g *AutoIncrementStringGenerator) SampleFromExisting(r *rand.Rand) interface{} {
	return g.SampleFromExistingTyped(r)
}

func (g *AutoIncrementStringGenerator) GenerateTyped(r *rand.Rand) string {
	n := g.current.Add(1)
	return fmt.Sprintf("%d!%s", n, generateUniqueStringFromInt(n, g.length))
}

func (g *AutoIncrementStringGenerator) SampleFromExistingTyped(r *rand.Rand) string {
	n := r.Int63n(g.current.Load()-g.min+1) + g.min
	return generateUniqueStringFromInt(n, g.length)
}

func (g *AutoIncrementStringGenerator) Current() string {
	n := g.current.Load()
	return fmt.Sprintf("%d!%s", n, generateUniqueStringFromInt(n, g.length))
}

// Generates values from a discrete set of possible values.
//
// Sample from existing is the exact same as generation, which means it is
// possible to generate values not in the database but available in the set of
// values.
type EnumGenerator[T any] struct {
	values []T
}

func NewEnumGenerator[T any](values []T) *EnumGenerator[T] {
	return &EnumGenerator[T]{
		values: values,
	}
}

func (g *EnumGenerator[T]) Generate(r *rand.Rand) interface{} {
	return g.GenerateTyped(r)
}

func (g *EnumGenerator[T]) SampleFromExisting(r *rand.Rand) interface{} {
	return g.SampleFromExistingTyped(r)
}

func (g *EnumGenerator[T]) GenerateTyped(r *rand.Rand) T {
	return g.values[r.Intn(len(g.values))]
}

func (g *EnumGenerator[T]) SampleFromExistingTyped(r *rand.Rand) T {
	return g.GenerateTyped(r)
}

// Generates a datetime value that corresponds to the now time.
//
// Sample from existing does not work with this field and will panic.
type NowGenerator struct{}

func NewNowGenerator() NowGenerator {
	return NowGenerator{}
}

func (g NowGenerator) Generate(r *rand.Rand) interface{} {
	return g.GenerateTyped(r).Format("2006-01-02 15:04:05")
}

func (NowGenerator) SampleFromExisting(r *rand.Rand) interface{} {
	panic("cannot sample from existing with the now generator")
}

func (NowGenerator) GenerateTyped(r *rand.Rand) time.Time {
	return time.Now()
}

func (NowGenerator) SampleFromExistingTyped(r *rand.Rand) time.Time {
	panic("cannot sample from existing with the now generator")
}

// TODO: better date time generator

// Generates the same JSON document every time. This is based on
// map[string]string.
type JSONGenerator struct {
	objLength   int
	valueLength int
}

func NewJSONGenerator(objLength, valueLength int) *JSONGenerator {
	return &JSONGenerator{objLength: objLength, valueLength: valueLength}
}

func (g *JSONGenerator) Generate(r *rand.Rand) interface{} {
	return g.GenerateTyped(r)
}

func (g *JSONGenerator) SampleFromExisting(r *rand.Rand) interface{} {
	return g.SampleFromExistingTyped(r)
}

func (g *JSONGenerator) GenerateTyped(r *rand.Rand) string {
	m := make(map[string]string)

	for i := 0; i < g.objLength; i++ {
		v := strconv.Itoa(i)
		m[v] = strings.Repeat(v, g.valueLength)
	}

	data, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}

	return string(data)
}

func (g *JSONGenerator) SampleFromExistingTyped(r *rand.Rand) string {
	return g.GenerateTyped(r)
}

func generateUniqueStringFromInt(v int64, length int) string {
	// Add a dash at the end so that numbers like 1 and 11 will generate different
	// sequences. This is kind of a hack. No analysis was done to see if the
	// cardinality of the resulting distribution is the same as the input.
	//
	// TODO: add a test for this method and ensure cardinality?
	num := fmt.Sprintf("%d-", v)
	numIdx := 0

	var buf strings.Builder
	for i := 0; i < length; i++ {
		buf.WriteByte(num[numIdx])

		numIdx = (numIdx + 1) % len(num)
	}

	return buf.String()
}
