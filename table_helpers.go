package mybench

import (
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

func QuestionMarksStringList(n int) string {
	s := strings.Repeat("?,", n)
	return s[:len(s)-1]
}

type Column struct {
	// Name of the column
	Name string

	// SQL definition of the column
	Definition string

	// The data generator for the data of this column.
	Generator DataGenerator
}

// This struct provides helpers for creating and seeding a table.
type Table struct {
	// The name of the table
	Name string

	// The list of columns. The order of the columns in the table follows the
	// order of this slice.
	Columns []*Column

	// The columns for the primary key for the table. This is a slice as it
	// supports a composite primary key.
	PrimaryKey []string

	// A list of columns for indices.
	Indices [][]string

	// A list of columns for unique indices.
	UniqueKeys [][]string

	// Additional table options appended to the end of the CREATE TABLE
	// statements, such as compression settings, auto increment settings, and so
	// on.
	TableOptions string

	columnsMap map[string]*Column
}

func InitializeTable(t Table) Table {
	t.columnsMap = make(map[string]*Column)
	for _, column := range t.Columns {
		t.columnsMap[column.Name] = column
	}

	return t
}

func (t Table) Generate(r *Rand, column string) interface{} {
	return t.columnsMap[column].Generator.Generate(r)
}

func (t Table) SampleFromExisting(r *Rand, column string) interface{} {
	return t.columnsMap[column].Generator.SampleFromExisting(r)
}

func (t Table) CreateTableQuery() string {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("CREATE TABLE `%s` (", t.Name))

	for _, column := range t.Columns {
		buf.WriteString(fmt.Sprintf("`%s` %s", column.Name, column.Definition))
		buf.WriteString(",")
	}

	buf.WriteString("PRIMARY KEY (")
	for i, column := range t.PrimaryKey {
		buf.WriteString(fmt.Sprintf("`%s`", column))
		if i < len(t.PrimaryKey)-1 {
			buf.WriteString(",")
		}
	}
	buf.WriteString(")")

	if len(t.Indices) > 0 {
		buf.WriteString(",")
		for i, index := range t.Indices {
			buf.WriteString(fmt.Sprintf("KEY (%s)", strings.Join(index, ",")))
			if i < len(t.Indices)-1 {
				buf.WriteString(",")
			}
		}
	}

	if len(t.UniqueKeys) > 0 {
		buf.WriteString(",")
		for i, index := range t.UniqueKeys {
			buf.WriteString(fmt.Sprintf("UNIQUE KEY (%s)", strings.Join(index, ",")))
			if i < len(t.UniqueKeys)-1 {
				buf.WriteString(",")
			}
		}
	}

	buf.WriteString(")")

	if t.TableOptions != "" {
		buf.WriteString(" ")
		buf.WriteString(t.TableOptions)
	}

	return buf.String()
}

func (t Table) DropTableQuery() string {
	return fmt.Sprintf("DROP TABLE IF EXISTS `%s`", t.Name)
}

func (t Table) InsertQuery(r *Rand, batchSize int, valueOverride map[string]interface{}) (string, []interface{}) {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("INSERT INTO `%s` (", t.Name))
	for i, column := range t.Columns {
		buf.WriteString("`" + column.Name + "`")
		if i < len(t.Columns)-1 {
			buf.WriteString(",")
		}
	}
	buf.WriteString(") VALUES ")

	questionMarks := strings.Repeat("?,", len(t.Columns))
	questionMarks = "(" + questionMarks[:len(questionMarks)-1] + "),"

	questionMarks = strings.Repeat(questionMarks, batchSize)
	questionMarks = questionMarks[:len(questionMarks)-1]
	buf.WriteString(questionMarks)

	args := make([]interface{}, 0, len(t.Columns)*batchSize)
	for i := 0; i < batchSize; i++ {
		for _, column := range t.Columns {
			var value interface{}
			if valueOverride != nil {
				v, found := valueOverride[column.Name]
				if found {
					value = v
				} else {
					value = column.Generator.Generate(r)
				}
			} else {
				value = column.Generator.Generate(r)
			}

			args = append(args, value)
		}
	}

	return buf.String(), args
}

func (t Table) InsertQueryList(r *Rand, valueOverrides []map[string]interface{}) (string, []interface{}) {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("INSERT INTO `%s` (", t.Name))
	for i, column := range t.Columns {
		buf.WriteString("`" + column.Name + "`")
		if i < len(t.Columns)-1 {
			buf.WriteString(",")
		}
	}
	buf.WriteString(") VALUES ")

	questionMarks := strings.Repeat("?,", len(t.Columns))
	questionMarks = "(" + questionMarks[:len(questionMarks)-1] + "),"

	questionMarks = strings.Repeat(questionMarks, len(valueOverrides))
	questionMarks = questionMarks[:len(questionMarks)-1]
	buf.WriteString(questionMarks)

	args := make([]interface{}, 0, len(t.Columns)*len(valueOverrides))
	for _, valueOverride := range valueOverrides {
		for _, column := range t.Columns {
			var value interface{}
			if valueOverride != nil {
				v, found := valueOverride[column.Name]
				if found {
					value = v
				} else {
					value = column.Generator.Generate(r)
				}
			} else {
				value = column.Generator.Generate(r)
			}

			args = append(args, value)
		}
	}

	return buf.String(), args
}

// Drop and recreate the table with data seeded via the data generators.
//
// totalrows specifies the total number of rows to insert into the new table.
// batchSize controls how many rows to insert in one INSERT statement (200 is
// usually a good starting point), concurrency is the number of goroutines used
// to insert the data.
//
// If concurrency is 0, it is set by default to 16. This allows the loader to
// reuse the -concurrency flag (which is default 0).
func (t Table) ReloadData(databaseConfig DatabaseConfig, totalrows int64, batchSize int64, concurrency int) {
	if concurrency <= 0 { // apply default if no valid concurrency is given
		concurrency = 16
	}

	logger := logrus.WithFields(logrus.Fields{
		"table":     t.Name,
		"totalrows": totalrows,
	})

	logger.WithFields(logrus.Fields{
		"batchSize":   batchSize,
		"concurrency": concurrency,
	}).Info("reloading data")

	conn, err := databaseConfig.Connection()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	_, err = conn.Execute(t.DropTableQuery())
	if err != nil {
		panic(err)
	}

	_, err = conn.Execute(t.CreateTableQuery())
	if err != nil {
		panic(err)
	}

	wg := &sync.WaitGroup{}
	batchSizeChan := make(chan int64)

	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()

			conn, err := databaseConfig.Connection()
			if err != nil {
				panic(err)
			}
			defer conn.Close()

			r := NewRand()

			for {
				batchSize, open := <-batchSizeChan
				if !open {
					return
				}

				query, args := t.InsertQuery(r, int(batchSize), nil)
				_, err = conn.Execute(query, args...)
				if err != nil {
					logger.WithFields(logrus.Fields{
						"query": query,
						"args":  args,
					}).Error(err)
					panic(err)
				}
			}
		}()
	}

	rowsInserted := int64(0)
	lastLoggedPct := -1.0

	for rowsInserted < totalrows {
		if totalrows-rowsInserted < batchSize {
			batchSize = totalrows - rowsInserted
		}

		pct := float64(rowsInserted) / float64(totalrows) * 100
		if pct-lastLoggedPct > 1 {
			lastLoggedPct = pct
			logger.WithFields(logrus.Fields{"pct": math.Round(pct*100) / 100.0, "rowsInserted": rowsInserted}).Info("loading data")
		}

		batchSizeChan <- batchSize
		rowsInserted += batchSize
	}

	close(batchSizeChan)
	wg.Wait()
	logger.WithFields(logrus.Fields{"pct": 100.0, "rowsInserted": rowsInserted}).Info("data reloaded")
}
