package schema

import (
	"encoding/xml"
	"strings"

	"github.com/beevik/etree"
)

// ================================
// Structure for uschema
// ================================

type USchema struct {
	XMLName      xml.Name `xml:"USchema:USchema"`
	XmlnsXmi     string   `xml:"xmlns:xmi,attr"`
	XmlnsXsi     string   `xml:"xmlns:xsi,attr"`
	XmlnsUSchema string   `xml:"xmlns:USchema,attr"`
	Name         string   `xml:"name,attr"`
	Entities     []Entity `xml:"entities"`
}

type Entity struct {
	Name       string      `xml:"name,attr"`
	Root       string      `xml:"root,attr,omitempty"`
	Variations []Variation `xml:"variations"`
}

type Variation struct {
	VariationId        string    `xml:"variationId,attr"`
	Count              string    `xml:"count,attr"`
	LogicalFeatures    string    `xml:"logicalFeatures,attr,omitempty"`
	StructuralFeatures string    `xml:"structuralFeatures,attr,omitempty"`
	Features           []Feature `xml:"features"`
}

type Feature struct {
	XsiType    string `xml:"xsi:type,attr"`
	Name       string `xml:"name,attr"`
	RefsTo     string `xml:"refsTo,attr,omitempty"`
	Key        string `xml:"key,attr,omitempty"`
	Attributes string `xml:"attributes,attr,omitempty"`
}

func ExtractLabels(entity *etree.Element) []string {
	name := entity.SelectAttrValue("name", "")
	if name == "" {
		return []string{}
	}
	return strings.Split(name, "_AND_")
}
