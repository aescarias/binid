package bindef

import (
	"fmt"
	"strconv"
	"strings"
)

// SpecVersion is the version of the BinDef spec implemented by this runtime.
var SpecVersion = Version{Major: 0, Minor: 1}

// A Version describes a specification version with a major and minor version component.
type Version struct {
	Major int
	Minor int
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

// NewVersionFromString produces a Version struct from a string of the form
// "major.minor". If this parsing fails, an error is returned.
func NewVersionFromString(version string) (Version, error) {
	split := strings.SplitN(version, ".", 2)

	major, err := strconv.Atoi(split[0])
	if err != nil {
		return Version{}, err
	}

	minor, err := strconv.Atoi(split[1])
	if err != nil {
		return Version{}, err
	}

	return Version{Major: major, Minor: minor}, nil
}

func ident(value string) Result {
	return Result{Kind: ResultIdent, Value: value}
}

type Meta struct {
	BdfVersion Version
	Name       string
	Mime       []string
	Exts       []string
}

func parseMeta(meta Result) (Meta, error) {
	if meta.Kind != ResultMap {
		return Meta{}, fmt.Errorf("meta must be a mapping")
	}

	metadata := meta.Value.(map[Result]Result)
	bdf, ok := metadata[ident("bdf")]
	if !ok {
		return Meta{}, fmt.Errorf("'meta' map missing required key 'bdf'")
	}

	if bdf.Kind != ResultString {
		return Meta{}, fmt.Errorf("meta.bdf must be a constant string")
	}

	version, err := NewVersionFromString(bdf.Value.(string))
	if err != nil {
		return Meta{}, err
	}

	// name
	name, ok := metadata[ident("name")]
	if !ok {
		return Meta{}, fmt.Errorf("'meta' map missing required key 'name'")
	}

	if name.Kind != ResultString {
		return Meta{}, fmt.Errorf("meta.name must be a constant string")
	}

	// mime
	mime, ok := metadata[ident("mime")]
	if !ok {
		return Meta{}, fmt.Errorf("'meta' map missing required key 'mime'")
	}

	if mime.Kind != ResultList {
		return Meta{}, fmt.Errorf("meta.mime must be a constant list")
	}

	mimeRes := mime.Value.([]Result)
	mimeList := make([]string, len(mimeRes))

	for idx, item := range mimeRes {
		if item.Kind != ResultString {
			return Meta{}, fmt.Errorf("meta.mime: element at index %d is not a string", idx)
		}

		mimeList[idx] = item.Value.(string)
	}

	// exts
	exts, ok := metadata[ident("exts")]
	if !ok {
		return Meta{}, fmt.Errorf("'meta' map missing required key 'exts'")
	}

	if exts.Kind != ResultList {
		return Meta{}, fmt.Errorf("meta.exts must be a constant list")
	}

	extsRes := exts.Value.([]Result)
	extsList := make([]string, len(extsRes))

	for idx, item := range extsRes {
		if item.Kind != ResultString {
			return Meta{}, fmt.Errorf("meta.exts: element at index %d is not a string", idx)
		}

		extsList[idx] = item.Value.(string)
	}

	return Meta{
		BdfVersion: version,
		Name:       name.Value.(string),
		Mime:       mimeList,
		Exts:       extsList,
	}, nil
}

// GetMetadata returns the metadata described in the 'meta' key of document.
func GetMetadata(document Result) (Meta, error) {
	if document.Kind != ResultMap {
		return Meta{}, fmt.Errorf("document must be a mapping")
	}

	mapping := document.Value.(map[Result]Result)
	meta, err := parseMeta(mapping[ident("meta")])

	return meta, err
}
