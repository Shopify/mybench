package mybench

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/atomic"
)

const characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// An interface for the data generator.
//
// There are two ways to generate data:
//
// 1. Generate a new value to be inserted into the database. This is generated
// via the Generate call.
// 2. Generate an "existing" value to be used in the WHERE clause of a SQL
// statement. This is generated via the SampleFromExisting call. Note, most
// generators cannot guarantee that an existing value is generated, as it would
// be probihitively expensive to keep track of all the existing data. Consult
// the documentation of the specific generators for details.
type DataGenerator interface {
	Generate(*Rand) interface{}
	SampleFromExisting(*Rand) interface{}
}

// A boring generator that only generates only null values.
type NullGenerator struct{}

func NewNullGenerator() NullGenerator {
	return NullGenerator{}
}

func (NullGenerator) Generate(*Rand) interface{} {
	return nil
}

func (NullGenerator) SampleFromExisting(*Rand) interface{} {
	return nil
}

// =================
// Number generators
// =================

// Generates an integer value in the inclusive range between min (inclusive)
// and max (exclusive) with an uniform distribution.
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

func (g *UniformIntGenerator) Generate(r *Rand) interface{} {
	return g.GenerateTyped(r)
}

func (g *UniformIntGenerator) SampleFromExisting(r *Rand) interface{} {
	return g.SampleFromExistingTyped(r)
}

func (g *UniformIntGenerator) GenerateTyped(r *Rand) int64 {
	return r.UniformInt(g.min, g.max)
}

func (g *UniformIntGenerator) SampleFromExistingTyped(r *Rand) int64 {
	return g.GenerateTyped(r)
}

// Generates a random floating point value according to an uniform distribution
// between min (inclusive) and max (exclusive).
//
// Sampling from existing is the same as the generation, since there are a
// large number of floating point values, it is unlikely to generate an exact
// value that has been used before. However, the generated value may still be
// useful in WHERE clauses that uses the greater than or less than operators.
type UniformFloatGenerator struct {
	min float64
	max float64
}

func NewUniformFloatGenerator(min, max float64) *UniformFloatGenerator {
	return &UniformFloatGenerator{min, max}
}

func (g *UniformFloatGenerator) Generate(r *Rand) interface{} {
	return g.GenerateTyped(r)
}

func (g *UniformFloatGenerator) GenerateTyped(r *Rand) float64 {
	return r.UniformFloat(g.min, g.max)
}

func (g *UniformFloatGenerator) SampleFromExisting(r *Rand) interface{} {
	return g.SampleFromExistingTyped(r)
}

func (g *UniformFloatGenerator) SampleFromExistingTyped(r *Rand) float64 {
	return g.GenerateTyped(r)
}

// Generates a random integer value according to a normal distribution.
//
// Sample from existing is the same as generation.
type NormalIntGenerator struct {
	mean   int64
	stddev int64
}

func NewNormalIntGenerator(mean, stddev int64) *NormalIntGenerator {
	return &NormalIntGenerator{
		mean:   mean,
		stddev: stddev,
	}
}

func (g *NormalIntGenerator) Generate(r *Rand) interface{} {
	return g.GenerateTyped(r)
}

func (g *NormalIntGenerator) GenerateTyped(r *Rand) int64 {
	return r.NormalInt(g.mean, g.stddev)
}

func (g *NormalIntGenerator) SampleFromExisting(r *Rand) interface{} {
	return g.SampleFromExistingTyped(r)
}

func (g *NormalIntGenerator) SampleFromExistingTyped(r *Rand) int64 {
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

func (g *NormalFloatGenerator) Generate(r *Rand) interface{} {
	return g.GenerateTyped(r)
}

func (g *NormalFloatGenerator) SampleFromExisting(r *Rand) interface{} {
	return g.SampleFromExistingTyped(r)
}

func (g *NormalFloatGenerator) GenerateTyped(r *Rand) float64 {
	return r.NormFloat64()*g.stddev + g.mean
}

func (g *NormalFloatGenerator) SampleFromExistingTyped(r *Rand) float64 {
	return g.GenerateTyped(r)
}

// Generates integers according to a histogram distribution. One possible use
// case of this is when you want to distribute a foreign key/id with a
// particular distribution. For example, a `posts` table can have many posts,
// with 50% of the rows having one `user_id`, and then 25% of the rows with
// another `user_id`.
//
// Sample from existing does not track of values already generated but samples
// from the same distribution as the Generate. This means it is possible to
// generate values that doesn't exist in the database.
type HistogramIntGenerator struct {
	hist HistogramDistribution
}

// See NewHistogramDistribution for documentation the arguments for this
// function. Note each integer generated by the histogram will be mapped to a
// string. To specify make sure integers such as 1, 2, 3, 4 are generated, the
// binsEndPoints must be 0.5, 1.5, 2.5, 3.5, 4.5.
func NewHistogramIntGenerator(binsEndPoints, frequency []float64) *HistogramIntGenerator {
	return &HistogramIntGenerator{
		hist: NewHistogramDistribution(binsEndPoints, frequency),
	}
}

func (g *HistogramIntGenerator) Generate(r *Rand) interface{} {
	return g.GenerateTyped(r)
}

func (g *HistogramIntGenerator) GenerateTyped(r *Rand) int64 {
	return r.HistInt(g.hist)
}

func (g *HistogramIntGenerator) SampleFromExisting(r *Rand) interface{} {
	return g.SampleFromExistingTyped(r)
}

func (g *HistogramIntGenerator) SampleFromExistingTyped(r *Rand) int64 {
	return g.GenerateTyped(r)
}

// Generates floating point values according to a histogram distribution.
//
// Sample from existing does not track of values already generated but samples
// from the same distribution as the Generate. This means it is possible to
// generate values that doesn't exist in the database.
type HistogramFloatGenerator struct {
	hist HistogramDistribution
}

// See NewHistogramDistribution for documentation the arguments for this
// function.
func NewHistogramFloatGenerator(binsEndPoints, frequency []float64) *HistogramFloatGenerator {
	return &HistogramFloatGenerator{
		hist: NewHistogramDistribution(binsEndPoints, frequency),
	}
}

func (g *HistogramFloatGenerator) Generate(r *Rand) interface{} {
	return g.GenerateTyped(r)
}

func (g *HistogramFloatGenerator) GenerateTyped(r *Rand) float64 {
	return r.HistFloat(g.hist)
}

func (g *HistogramFloatGenerator) SampleFromExisting(r *Rand) interface{} {
	return g.SampleFromExistingTyped(r)
}

func (g *HistogramFloatGenerator) SampleFromExistingTyped(r *Rand) float64 {
	return g.GenerateTyped(r)
}

// TODO: can this be folded into the UniformFloatGenerator?
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

func (g *UniformDecimalGenerator) Generate(r *Rand) interface{} {
	return g.GenerateTyped(r)
}

func (g *UniformDecimalGenerator) SampleFromExisting(r *Rand) interface{} {
	return g.SampleFromExistingTyped(r)
}

func (g *UniformDecimalGenerator) GenerateTyped(r *Rand) string {
	num := rand.Float64() * math.Pow10(g.precision) / math.Pow10(g.scale)
	format := fmt.Sprintf("%%%d.%df", g.precision, g.scale)
	return fmt.Sprintf(format, num)
}

func (g *UniformDecimalGenerator) SampleFromExistingTyped(r *Rand) string {
	return g.GenerateTyped(r)
}

// ================
// String generator
// ================

// Generates a fixed number of unique strings with uniform distribution. For
// example, if cardinality is 10, then this generator will generate 10 distinct
// string values. The frequency of the strings are uniform.
//
// Sample from existing is the same as generate, which means it may not sample
// an existing value.
type UniformCardinalityStringGenerator struct {
	cardinality int
	length      int
}

func NewUniformCardinalityStringGenerator(cardinality, length int) *UniformCardinalityStringGenerator {
	return &UniformCardinalityStringGenerator{cardinality: cardinality, length: length}
}

func (g *UniformCardinalityStringGenerator) Generate(r *Rand) interface{} {
	return g.GenerateTyped(r)
}

func (g *UniformCardinalityStringGenerator) GenerateTyped(r *Rand) string {
	i := r.UniformInt(0, int64(g.cardinality))
	return generateUniqueStringFromInt(i, g.length)
}

func (g *UniformCardinalityStringGenerator) SampleFromExisting(r *Rand) interface{} {
	return g.SampleFromExistingTyped(r)
}

func (g *UniformCardinalityStringGenerator) SampleFromExistingTyped(r *Rand) string {
	return g.GenerateTyped(r)
}

// Generates a fixed number of unique strings with uniform distribution. For
// example, if cardinality is 10, then this generator will generate 10 distinct
// string values. The frequency of the strings are uniform.
//
// Sample from existing is the same as generate, which means it may not sample
// an existing value.
type HistogramCardinalityStringGenerator struct {
	hist   HistogramDistribution
	length int
}

// See NewHistogramDistribution for documentation the arguments for this
// function. Note each integer generated by the histogram will be mapped to a
// string. To specify make sure integers such as 1, 2, 3, 4 are generated, the
// binsEndPoints must be 0.5, 1.5, 2.5, 3.5, 4.5.
func NewHistogramCardinalityStringGenerator(binsEndPoints, frequency []float64, length int) *HistogramCardinalityStringGenerator {
	return &HistogramCardinalityStringGenerator{
		hist:   NewHistogramDistribution(binsEndPoints, frequency),
		length: length,
	}
}

func (g *HistogramCardinalityStringGenerator) Generate(r *Rand) interface{} {
	return g.GenerateTyped(r)
}

func (g *HistogramCardinalityStringGenerator) GenerateTyped(r *Rand) string {
	i := r.HistInt(g.hist)
	return generateUniqueStringFromInt(i, g.length)
}

func (g *HistogramCardinalityStringGenerator) SampleFromExisting(r *Rand) interface{} {
	return g.SampleFromExistingTyped(r)
}

func (g *HistogramCardinalityStringGenerator) SampleFromExistingTyped(r *Rand) string {
	return g.GenerateTyped(r)
}

// Generates a random string with length selected between the min and max
// specified with uniform probability.
//
// Sample from existing is the same as generate and does not keep track of
// existing values. Since there are a very large amount of possible strings
// being generated, there is almost no chance that an existing value will be
// generated. It is best not to use that method and expect good results.
type UniformLengthStringGenerator struct {
	minLength int
	maxLength int
}

func NewUniformLengthStringGenerator(minLength, maxLength int) *UniformLengthStringGenerator {
	return &UniformLengthStringGenerator{
		minLength: minLength,
		maxLength: maxLength,
	}
}

func (g *UniformLengthStringGenerator) Generate(r *Rand) interface{} {
	return g.GenerateTyped(r)
}

func (g *UniformLengthStringGenerator) GenerateTyped(r *Rand) string {
	length := int64(g.maxLength)
	if g.minLength < g.maxLength {
		length = r.UniformInt(int64(g.minLength), int64(g.maxLength))
	}
	buf := make([]byte, length)
	for i := int64(0); i < length; i++ {
		buf[i] = characters[r.Intn(len(characters))]
	}

	return string(buf)
}

func (g *UniformLengthStringGenerator) SampleFromExisting(r *Rand) interface{} {
	return g.SampleFromExistingTyped(r)
}

func (g *UniformLengthStringGenerator) SampleFromExistingTyped(r *Rand) string {
	return g.GenerateTyped(r)
}

// Generates a random string with length selected by a histogram distribution.
//
// Sample from existing is the same as generate and does not keep track of
// existing values. Since there are a very large amount of possible strings
// being generated, there is almost no chance that an existing value will be
// generated. It is best not to use that method and expect good results.
type HistogramLengthStringGenerator struct {
	hist HistogramDistribution
}

// See NewHistogramDistribution for documentation the arguments for this
// function. Note each integer generated by the histogram will be mapped to a
// string. To specify make sure integers such as 1, 2, 3, 4 are generated, the
// binsEndPoints must be 0.5, 1.5, 2.5, 3.5, 4.5.
func NewHistogramLengthStringGenerator(binsEndPoints, frequency []float64) *HistogramLengthStringGenerator {
	return &HistogramLengthStringGenerator{
		hist: NewHistogramDistribution(binsEndPoints, frequency),
	}
}

func (g *HistogramLengthStringGenerator) Generate(r *Rand) interface{} {
	return g.GenerateTyped(r)
}

func (g *HistogramLengthStringGenerator) GenerateTyped(r *Rand) string {
	length := r.HistInt(g.hist)
	buf := make([]byte, length)
	for i := int64(0); i < length; i++ {
		buf[i] = characters[r.Intn(len(characters))]
	}

	return string(buf)
}

func (g *HistogramLengthStringGenerator) SampleFromExisting(r *Rand) interface{} {
	return g.SampleFromExistingTyped(r)
}

func (g *HistogramLengthStringGenerator) SampleFromExistingTyped(r *Rand) string {
	return g.GenerateTyped(r)
}

// Generates an unique string with a fixed length every time Generate is
// called. The internal generation is based on an atomic, incrementing integer.
// Each integer is converted into a string (via a hash function).
//
// Sample from existing will generate a value that has previously been
// generated. However, the value may have been deleted in the database so it's
// not guaranteed that the value generated will exist on the database.
type UniqueStringGenerator struct {
	min     int64
	current *atomic.Int64
	length  int
}

// length is the length of the string to be generated
// min and current are the integer values used to generate the strings. For
// loading data (when there are nothing in the database), min and current both
// should be 0. When there are already data in the database, min and current
// should be set to the min and max integer values used to generate strings
// that already exist in the database.
func NewUniqueStringGenerator(length int, min, current int64) *UniqueStringGenerator {
	return &UniqueStringGenerator{
		min:     min,
		current: atomic.NewInt64(current),
		length:  length,
	}
}

func NewUniqueStringGeneratorFromDatabase(databaseConfig DatabaseConfig, table, column string) (*UniqueStringGenerator, error) {
	conn, err := databaseConfig.Connection()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	query := fmt.Sprintf("SELECT MIN(CAST(SUBSTRING_INDEX(%s, '!', 1) AS UNSIGNED)) AS min_value, MAX(CAST(SUBSTRING_INDEX(%s, '!', 1) AS UNSIGNED)) AS current_value FROM `%s`.`%s`", column, column, databaseConfig.Database, table)
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

	query = fmt.Sprintf("SELECT LENGTH(`%s`) FROM `%s`.`%s` LIMIT 1", column, databaseConfig.Database, table)
	res, err = conn.Execute(query)
	if err != nil {
		return nil, err
	}

	length, err := res.GetInt(0, 0)
	if err != nil {
		return nil, err
	}

	return &UniqueStringGenerator{
		min:     min,
		current: atomic.NewInt64(current),
		length:  int(length),
	}, nil
}

func (g *UniqueStringGenerator) Generate(r *Rand) interface{} {
	return g.GenerateTyped(r)
}

func (g *UniqueStringGenerator) GenerateTyped(r *Rand) string {
	i := g.current.Add(1)
	return generateUniqueStringFromInt(i, g.length)
}

func (g *UniqueStringGenerator) SampleFromExisting(r *Rand) interface{} {
	return g.SampleFromExistingTyped(r)
}

func (g *UniqueStringGenerator) SampleFromExistingTyped(r *Rand) string {
	max := g.current.Load()
	// The reason we add 1 is because the values we use to generate is always the
	// value post add rather than pre add. If we do not add 1, we will sample the
	// first value corresponding to min and will not sample the last value
	// corresponding to current.
	i := r.UniformInt(g.min+1, max+1)
	return generateUniqueStringFromInt(i, g.length)
}

type DatetimeInterval struct {
	Start time.Time
	End   time.Time
}

// Generates a date time value in two modes:
//
//  1. GenerateNow == true will cause Generate to return time.Now.
//  2. GenerateNow == false will cause Generate to generate a random time
//     between the intervals specified in Intervals with uniform probability
//     distribution.
//
// SampleFromExisting always will sample from the Intervals. However, if
// GenerateNow == true, then it will also sample between an extra interval
// between when Generate() is first called and the moment the
// SampleFromExisting call is made.
//
// Generate and SampleFromExisting will return a string of the time formatted
// with YYYY-MM-DD hh:mm:ss, which is what SQL expects. GenerateTyped and
// SampleFromExistingTyped will return time.Time.
type UniformDatetimeGenerator struct {
	// intervals where a datetime value will be randomly generated according to
	// an uniform distribution. SampleFromExisting will always generate values
	// with these intervals. Generate will use these intervals unless GenerateNow
	// is true. In that case, Generate will add an additional interval between
	// when the first Generate is called and time.Now.
	intervals []DatetimeInterval

	// Generate will always return Now instead of sampling from the Intervals
	// specified. SampleFromExisting will still sample from the Intervals above,
	// although an interval between when the generator is first used and now will
	// be added to the SampleFromExisting.
	generateNow bool

	// Need a once object to ensure that the call to initialize firstGenerateTime
	// happens once in a thread-safe way.
	//
	// Note, sync.Once is fast (using atomic mostly and mutex only at the
	// beginning), so it shouldn't slowdown the generator.
	firstGenerateTime time.Time
	firstGenerateOnce sync.Once
}

func NewNowGenerator() *UniformDatetimeGenerator {
	return NewUniformDatetimeGenerator(nil, true)
}

func NewUniformDatetimeGenerator(intervals []DatetimeInterval, generateNow bool) *UniformDatetimeGenerator {
	return &UniformDatetimeGenerator{
		intervals:         intervals,
		generateNow:       generateNow,
		firstGenerateOnce: sync.Once{},
	}
}

func (g *UniformDatetimeGenerator) Generate(r *Rand) interface{} {
	return g.GenerateTyped(r).Format("2006-01-02 15:04:05")
}

func (g *UniformDatetimeGenerator) GenerateTyped(r *Rand) time.Time {
	if g.generateNow {
		g.firstGenerateOnce.Do(func() {
			g.firstGenerateTime = time.Now().UTC()
		})

		return time.Now().UTC()
	}

	return g.SampleFromExistingTyped(r)
}

func (g *UniformDatetimeGenerator) SampleFromExisting(r *Rand) interface{} {
	return g.SampleFromExistingTyped(r).Format("2006-01-02 15:04:05")
}

func (g *UniformDatetimeGenerator) SampleFromExistingTyped(r *Rand) time.Time {
	// Need to find a random interval first, and take into consideration of first
	// generate time if generateNow is enabled (which effectively forms another
	// interval).
	n := len(g.intervals)
	if g.generateNow {
		n++
	}

	idx := r.Intn(n)
	var randomInterval DatetimeInterval
	if idx == len(g.intervals) {
		randomInterval = DatetimeInterval{Start: g.firstGenerateTime, End: time.Now().UTC()}
	} else {
		randomInterval = g.intervals[idx]
	}

	randomDurationSeconds := r.Float64() * randomInterval.End.Sub(randomInterval.Start).Seconds()
	randomDuration := time.Duration(randomDurationSeconds) * time.Second

	return randomInterval.Start.Add(randomDuration)
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

func (g *UuidGenerator) Generate(r *Rand) interface{} {
	return g.GenerateTyped(r)
}

func (g *UuidGenerator) SampleFromExisting(r *Rand) interface{} {
	return g.SampleFromExistingTyped(r)
}

func (g *UuidGenerator) GenerateTyped(r *Rand) string {
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

func (g *UuidGenerator) SampleFromExistingTyped(r *Rand) string {
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

	query := fmt.Sprintf("SELECT MIN(%s), MAX(%s) FROM`%s`.`%s`", column, column, databaseConfig.Database, table)
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

func (g *AutoIncrementGenerator) Generate(r *Rand) interface{} {
	return g.GenerateTyped(r)
}

func (g *AutoIncrementGenerator) SampleFromExisting(r *Rand) interface{} {
	return g.SampleFromExistingTyped(r)
}

func (g *AutoIncrementGenerator) GenerateTyped(r *Rand) int64 {
	return g.current.Add(1)
}

func (g *AutoIncrementGenerator) SampleFromExistingTyped(r *Rand) int64 {
	return r.Int63n(g.Current()-g.min+1) + g.min
}

// Get the current value without generating a new value.
func (g *AutoIncrementGenerator) Current() int64 {
	return g.current.Load()
}

func (g *AutoIncrementGenerator) Min() int64 {
	return g.min
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

func (g *EnumGenerator[T]) Generate(r *Rand) interface{} {
	return g.GenerateTyped(r)
}

func (g *EnumGenerator[T]) SampleFromExisting(r *Rand) interface{} {
	return g.SampleFromExistingTyped(r)
}

func (g *EnumGenerator[T]) GenerateTyped(r *Rand) T {
	return g.values[r.Intn(len(g.values))]
}

func (g *EnumGenerator[T]) SampleFromExistingTyped(r *Rand) T {
	return g.GenerateTyped(r)
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

func (g *JSONGenerator) Generate(r *Rand) interface{} {
	return g.GenerateTyped(r)
}

func (g *JSONGenerator) SampleFromExisting(r *Rand) interface{} {
	return g.SampleFromExistingTyped(r)
}

func (g *JSONGenerator) GenerateTyped(r *Rand) string {
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

func (g *JSONGenerator) SampleFromExistingTyped(r *Rand) string {
	return g.GenerateTyped(r)
}

func generateUniqueStringFromInt(v int64, length int) string {
	hash := md5.Sum([]byte(fmt.Sprintf("%d", v)))
	hashStr := hex.EncodeToString(hash[:])
	hashStr = fmt.Sprintf("%d!%s", v, hashStr)
	if length == len(hashStr) {
		return hashStr
	}

	if length < len(hashStr) {
		return hashStr[:length]
	}

	// If the required length is bigger than the hash length, then we need to
	// extend the hash. The idea is to repeat the hash separated with -.
	var extendedHashBuf strings.Builder
	extendedHashBuf.WriteString(hashStr)

	for extendedHashBuf.Len() < length {
		extendedHashBuf.WriteByte('-')
		extendedHashBuf.WriteString(hashStr)
	}

	// If the repeated hash is longer than the required length, we truncate it
	// once again.
	return extendedHashBuf.String()[:length]
}
