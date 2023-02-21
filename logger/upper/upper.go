package upper

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/upper/db/v4"
)

const (
	fmtLogSessID       = `Session ID:     %05d`
	fmtLogTxID         = `Transaction ID: %05d`
	fmtLogQuery        = `Query:          %s`
	fmtLogArgs         = `Arguments:      %#v`
	fmtLogRowsAffected = `Rows affected:  %d`
	fmtLogLastInsertID = `Last insert ID: %d`
	fmtLogError        = `Error:          %v`
	fmtLogStack        = `Stack:          %v`
	fmtLogTimeTaken    = `Time taken:     %0.5fs`
	fmtLogContext      = `Context:        %v`
)

var _ db.Logger = &Logger{}

// Logger logger
type Logger struct {
	*logrus.Logger
}

// New logger
func New(l *logrus.Logger) (log *Logger) {
	log = &Logger{
		l,
	}
	return
}

// Print print
func (l *Logger) Print(v ...interface{}) {
	for _, _v := range v {
		qs, ok := _v.(*db.QueryStatus)
		if !ok {
			l.Logger.Print(_v)
		} else {
			l.Logger.Print(QueryStatusString(qs))
		}
	}
	// l.Logger.Print(v...)
}

// QueryStatusString QueryStatus string
func QueryStatusString(q *db.QueryStatus) string {
	lines := make([]string, 0, 8)

	if q.SessID > 0 {
		lines = append(lines, fmt.Sprintf(fmtLogSessID, q.SessID))
	}

	if q.TxID > 0 {
		lines = append(lines, fmt.Sprintf(fmtLogTxID, q.TxID))
	}

	if query := q.RawQuery; query != "" {
		lines = append(lines, fmt.Sprintf(fmtLogQuery, q.Query()))
	}

	if len(q.Args) > 0 {
		lines = append(lines, fmt.Sprintf(fmtLogArgs, q.Args))
	}

	// if stack := q.Stack(); len(stack) > 0 {
	//     lines = append(lines, fmt.Sprintf(fmtLogStack, "\n\t"+strings.Join(stack, "\n\t")))
	// }

	if q.RowsAffected != nil {
		lines = append(lines, fmt.Sprintf(fmtLogRowsAffected, *q.RowsAffected))
	}
	if q.LastInsertID != nil {
		lines = append(lines, fmt.Sprintf(fmtLogLastInsertID, *q.LastInsertID))
	}

	if q.Err != nil {
		lines = append(lines, fmt.Sprintf(fmtLogError, q.Err))
	}

	lines = append(lines, fmt.Sprintf(fmtLogTimeTaken, float64(q.End.UnixNano()-q.Start.UnixNano())/float64(1e9)))

	if q.Context != nil {
		lines = append(lines, fmt.Sprintf(fmtLogContext, q.Context))
	}

	return "\t" + strings.Replace(strings.Join(lines, "\n"), "\n", "\n\t", -1) + "\n\n"
}
