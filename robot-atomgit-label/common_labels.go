package main

import "regexp"

var (
	commonLabelRegex       = regexp.MustCompile(`(?m)^/(kind|priority|sig|good)\s*(.*?)\s*$`)
	removeCommonLabelRegex = regexp.MustCompile(`(?m)^/remove-(kind|priority|sig|good)\s*(.*?)\s*$`)
)

func getMatchedLabels(comment string) ([]string, []string) {
	return parseLabels(comment, commonLabelRegex),
		parseLabels(comment, removeCommonLabelRegex)
}

func parseLabels(comment string, reg *regexp.Regexp) []string {
	var labels []string
	r := reg.FindAllStringSubmatch(comment, -1)
	for _, v := range r {
		if len(v) < 3 {
			continue
		}

		for _, item := range v[2:] {
			if v[1] == "good" {
				prefix := v[1]
				labels = append(labels, prefix+item)
				continue
			}
			prefix := v[1] + "/"
			labels = append(labels, prefix+item)
		}
	}

	return labels
}
