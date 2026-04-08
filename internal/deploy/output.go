package deploy

import (
	"fmt"
	"io"
)

// List writes deploy targets to w (one per line: id TAB displayName).
func List(w io.Writer) error {
	lines, err := ListTargetSummaries()
	if err != nil {
		return err
	}
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return nil
}

// Show writes full schema description for id.
func Show(w io.Writer, id string) error {
	s, err := LoadSchema(id)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w, s.Describe())
	return err
}
