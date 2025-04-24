package format

import (
	"github.com/e-gun/HipparchiaGoServer/internal/base/str"
	"strings"
)

func FormatAnnotations(l str.DbWorkline) string {
	// HGB sets these in herxrunner.go

	// 	metadatacategories = map[int]string{
	//		0:   "newauthor",
	//		1:   "newwork",
	//		2:   "workabbrev",
	//		3:   "authabbrev",
	//		97:  "region",
	//		98:  "city",
	//		99:  "notes",
	//		100: "date",
	//		101: "publicationinfo",
	//		102: "additionalpubinfo",
	//		103: "stillfurtherpubinfo",
	//		108: "provenance",
	//		114: "reprints",
	//		116: "unknownmetadata116",
	//		122: "documentnumber",
	//	}

	//showlabelfor := map[string]string{
	//	//"region": "region",
	//	//"city":   "city",
	//	//"date":   "date",
	//	"documentnumber": "#",
	//}

	categoriestouse := map[string]map[string]bool{
		"gr": {
			"notes":          true,
			"date":           true,
			"city":           true,
			"region":         true,
			"documentnumber": true,
		},
		"lt": {
			"notes":  true,
			"date":   true,
			"city":   true,
			"region": true,
		},
		"in": {
			"date":      true,
			"corrected": true,
			"altered":   true,
			"rectified": true,
			"alternate": true,
			"discarded": true,
		},
		"dp": {
			"date":       true,
			"provenance": true,
			"corrected":  true,
			"altered":    true,
			"alternate":  true,
			"discarded":  true,
		},
		"ch": {
			"date":      true,
			"corrected": true,
			"altered":   true,
			"rectified": true,
			"alternate": true,
			"discarded": true,
		},
	}

	notemap := l.GatherMetadata() // map[string]string : { "category": "data" }

	//if len(notemap) != 0 {
	//	fmt.Println("notemap", notemap)
	//}

	touse := categoriestouse[l.GetCorpus()]

	var nn []string
	for k, v := range notemap {
		_, ok := touse[k]
		if ok {
			nn = append(nn, v)
		}
	}

	_, all := touse["all"]
	if all {
		for _, v := range notemap {
			nn = append(nn, v)
		}
	}

	simplenotes := l.GatherSimpleAnnotations()
	if simplenotes != "" {
		nn = append(nn, simplenotes)
	}

	return strings.Join(nn, "; ")
}
