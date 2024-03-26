//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package db

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/vv"
	"github.com/e-gun/wego/pkg/embedding"
	"github.com/jackc/pgx/v5"
	"io"
	"strings"
)

// VectorDBInitNN - initialize vv.VECTORTABLENAMENN
func VectorDBInitNN() {
	const (
		CREATE = `
			CREATE TABLE %s
			(
			  fingerprint character(32),
			  vectorsize  int,
			  vectordata  bytea
			)`
		EXISTS = "already exists"
	)
	ex := fmt.Sprintf(CREATE, vv.VECTORTABLENAMENN)
	_, err := SQLPool.Exec(context.Background(), ex)
	if err != nil {
		m := err.Error()
		if !strings.Contains(m, EXISTS) {
			Msg.EC(err)
		}
	} else {
		Msg.FYI("VectorDBInitNN(): success")
	}
}

// VectorDBCheckNN - has a search with this fingerprint already been stored?
func VectorDBCheckNN(fp string) bool {
	const (
		Q   = `SELECT fingerprint FROM %s WHERE fingerprint = '%s' LIMIT 1`
		F   = `VectorDBCheckNN() found %s`
		DNE = "does not exist"
	)

	q := fmt.Sprintf(Q, vv.VECTORTABLENAMENN, fp)
	foundrow, err := SQLPool.Query(context.Background(), q)
	if err != nil {
		m := err.Error()
		if strings.Contains(m, DNE) {
			VectorDBInitNN()
		}
		return false
	}

	type simplestring struct {
		S string
	}

	ss, err := pgx.CollectOneRow(foundrow, pgx.RowToStructByPos[simplestring])
	if err != nil {
		// mm := err.Error()
		// mm will be "no rows in result set" if you did not find the fingerprint
		return false
	} else {
		Msg.TMI(fmt.Sprintf(F, ss.S))
		return true
	}
}

// VectorDBAddNN - add a set of embeddings to vv.VECTORTABLENAMENN
func VectorDBAddNN(fp string, embs embedding.Embeddings) {
	const (
		MSG1 = "VectorDBAddNN(): "
		MSG2 = "%s compression: %dM -> %dM (-> %.1f%%)"
		MSG3 = "VectorDBAddNN() was sent empty embeddings"
		FAIL = "VectorDBAddNN() failed when calling json.Marshal(embs): nothing stored"
		INS  = `
			INSERT INTO %s
				(fingerprint, vectorsize, vectordata)
			VALUES ('%s', $1, $2)`
		GZ = gzip.BestSpeed
	)

	if embs.Empty() {
		Msg.PEEK(MSG3)
		return
	}

	// json vs jsi: jsoniter.ConfigFastest, this will marshal the float with 6 digits precision (lossy)
	eb, err := json.Marshal(embs)
	if err != nil {
		Msg.NOTE(FAIL)
		eb = []byte{}
	}

	// https://stackoverflow.com/questions/61077668/how-to-gzip-string-and-return-byte-array-in-golang
	var buf bytes.Buffer
	zw, err := gzip.NewWriterLevel(&buf, GZ)
	Msg.EC(err)
	_, err = zw.Write(eb)
	Msg.EC(err)
	err = zw.Close()
	Msg.EC(err)

	b := buf.Bytes()
	l2 := len(b)

	ex := fmt.Sprintf(INS, vv.VECTORTABLENAMENN, fp)

	_, err = SQLPool.Exec(context.Background(), ex, l2, b)
	Msg.EC(err)
	Msg.TMI(MSG1 + fp)

	// compressed is c. 33% of original
	// l1 := len(eb)
	// mm(fmt.Sprintf(MSG2, fp, l1/1024/1024, l2/1024/1024, (float32(l2)/float32(l1))*100), MSGTMI)
	buf.Reset()
}

// VectorDBFetchNN - get a set of embeddings from vv.VECTORTABLENAMENN
func VectorDBFetchNN(fp string) embedding.Embeddings {
	const (
		MSG1 = "VectorDBFetchNN(): "
		MSG2 = "VectorDBFetchNN() pulled empty set of embeddings for %s"
		Q    = `SELECT vectordata FROM %s WHERE fingerprint = '%s' LIMIT 1`
	)

	q := fmt.Sprintf(Q, vv.VECTORTABLENAMENN, fp)
	var vect []byte
	foundrow, err := SQLPool.Query(context.Background(), q)
	Msg.EC(err)

	defer foundrow.Close()
	for foundrow.Next() {
		err = foundrow.Scan(&vect)
		Msg.EC(err)
	}

	var buf bytes.Buffer
	buf.Write(vect)

	// the data in the tables is zipped and needs unzipping
	zr, err := gzip.NewReader(&buf)
	Msg.EC(err)
	err = zr.Close()
	Msg.EC(err)
	decompr, err := io.ReadAll(zr)
	Msg.EC(err)

	var emb embedding.Embeddings
	err = json.Unmarshal(decompr, &emb)
	Msg.EC(err)
	buf.Reset()

	if emb.Empty() {
		Msg.NOTE(fmt.Sprintf(MSG2, fp))
	}

	// mm(MSG1+fp, MSGPEEK)

	return emb
}

// VectorDBReset - drop vv.VECTORTABLENAMENN
func VectorDBReset() {
	const (
		MSG1 = "VectorDBReset() dropped "
		MSG2 = "VectorDBReset(): 'DROP TABLE %s' returned an (ignored) error: \n\t%s"
		E    = `DROP TABLE %s`
	)
	ex := fmt.Sprintf(E, vv.VECTORTABLENAMENN)

	_, err := SQLPool.Exec(context.Background(), ex)
	if err != nil {
		ms := err.Error()
		Msg.TMI(fmt.Sprintf(MSG2, vv.VECTORTABLENAMENN, ms))
	} else {
		Msg.NOTE(MSG1 + vv.VECTORTABLENAMENN)
	}
}

// VectorDBSizeNN - how much space is the vectordb using?
func VectorDBSizeNN(priority int) {
	const (
		SZQ  = "SELECT SUM(vectorsize) AS total FROM " + vv.VECTORTABLENAMENN
		MSG4 = "Disk space used by stored vectors is currently %dMB"
	)
	var size int64

	err := SQLPool.QueryRow(context.Background(), SZQ).Scan(&size)
	Msg.EC(err)
	Msg.Emit(fmt.Sprintf(MSG4, size/1024/1024), priority)
}

func VectorDBCountNN(priority int) {
	const (
		SZQ  = "SELECT COUNT(vectorsize) AS total FROM " + vv.VECTORTABLENAMENN
		MSG4 = "Number of stored vector models: %d"
		DNE  = "does not exist"
	)
	var size int64

	err := SQLPool.QueryRow(context.Background(), SZQ).Scan(&size)
	if err != nil {
		m := err.Error()
		if strings.Contains(m, DNE) {
			VectorDBInitNN()
		}
		size = 0
	}
	Msg.Emit(fmt.Sprintf(MSG4, size), priority)
}
