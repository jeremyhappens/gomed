package gomed


/*
Copyright 2012 Jeremy Sullivan

This file is part of gomed.

gomed is free software: you can redistribute it and/or modify
it under the terms of the GNU Lesser General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

gomed is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Lesser General Public License for more details.

You should have received a copy of the GNU Lesser General Public License
along with gomed.  If not, see <http://www.gnu.org/licenses/>.
*/


import (
  "fmt"
  "strings"
  "encoding/xml"
  "os"
  "strconv"
  "bytes"
  //"net"
  "xmlx"
  "io"
)
/*

Marshall HL7 2.x into xml (not hl7 3.x) and back again.

The parsing functions are experimental at best and haven't been exposed to a wide sampling of
hl7.

TODO:
This won't parse anything besides the standard seperators (| and ^) and utilizes \r as segment seperators.
It doesn't add anything to indicate the start or stop of a message.
To that end, it should probably toss out any other nonword characters.

This is meant to be exposed to the network or database packages so it can work with external interfaces.
This was never meant to be used as a stand alone HL7 tool.
*/

//Some string splitting functions.
func segs(msg string) []string {
	return strings.Split(msg, "\r")
}

func fieldSplit(segment string) []string {
  s := strings.Split(segment, "|")
  for i := 0; i < len(s); i++ {
    if(strings.Contains(s[i], "&")){
      s[i] = strings.Replace(s[i], "&", "&amp;", -1)
     }
    }
    return s
}

//some structs used in xml marshalling
type Hl7Doc struct {
	XMLName xml.Name
	Header []Segment
}

type Segment struct {
	XMLName xml.Name
	Fields []SegmentField
}

type SegmentField struct {
	XMLName xml.Name
	Repeating []RepField
	Data string `xml:",innerxml"`
}

type RepField struct {
	XMLName xml.Name
	Data string `xml:",innerxml"`
}


//a collection of really bad struct constructors
func newRepField(repnum int, datastring string) RepField {
  return RepField{XMLName: xml.Name{
    Space: "", Local: "CE" + "." + strconv.Itoa(repnum)},
    Data: datastring }
}

func newSegField(header string, repnum int, datastring string) SegmentField {
  return SegmentField{
    XMLName: xml.Name{
      Space: "",
      Local: header + "." + strconv.Itoa(repnum)},
    Data: datastring}
}

func newSegFields(header string, repnum int, reps []RepField) SegmentField {
  return SegmentField{
    XMLName: xml.Name{
      Space: "",
      Local: header + "." + strconv.Itoa(repnum)},
    Repeating: reps}
}


func toXML(msg string) Hl7Doc {
	segmentStrings := segs(msg)
	var segments []Segment
	var msg_type string
	
	for i := 0; i < len(segmentStrings); i++ {
		var fields []SegmentField
		fieldStrings := fieldSplit(segmentStrings[i])
		header := fieldStrings[0]
		
    //check if we're working with the header, if so take out the ^ and replace with _
		if(header == "MSH") {
			msg_type = fieldStrings[8]
      msg_type = strings.Replace(msg_type, "^", "_", 1)
		}

    //This really shouldn't happen. But if you're working with very broken hl7 it's possible.
    if(header == "") {
      continue
      header = "None"
    }

		//put together child fields
		for j := 0; j < len(fieldStrings); j++ {

      //XML marshalling really hates that &. So we need to clean it up.
      if(strings.Contains(fieldStrings[j], "&")){
        fieldStrings[j] = strings.Replace(fieldStrings[j], "&", "&amp;", 1)
      }

      //If we have a ^, it's a repeating field and we need to pile those into a slice
      //TODO: add support for a seperator other than ^
      //TODO: Specifically, get the character out of the header field definition.
			if(strings.Contains(fieldStrings[j], "^")){
				var reps []RepField
				subfields := strings.Split(fieldStrings[j], "^")
				for k := 0; k < len(subfields); k++ {
          reps = append(reps, newRepField(k, subfields[k]))
				}
        fields = append(fields, newSegFields(header, j, reps))
			} else {
        fields = append(fields, newSegField(header, j, fieldStrings[j]))
			}
		}
    //insert <Header.#> with [<field>data</field>]
		segments = append(segments, Segment{
			XMLName: xml.Name{
				Space: "",
				Local: header},
			Fields: fields})
	}
	
	document := Hl7Doc{XMLName: xml.Name{
			Space: "", Local: msg_type},
			Header: segments}
	
	return document
}

//TODO: Better error handling
func FromFile(filename string) Hl7Doc {
  file, ferr := os.Open(filename)

  if ferr != nil {
    //return it
    println(ferr)
  }

  fstats, serr := file.Stat()

  if serr != nil {
    println(serr)
  }

  data := make([]byte, fstats.Size())
  count, rerr := file.Read(data)

  if rerr != nil {
    println(rerr)
  }

  println(count)

  return toXML(string(data))


}

//this function is deprecated as toHl7 is deprecated. Kept around for reference
/*
func XmlFromFile(filename string) *bytes.Buffer {

  file, ferr := os.Open(filename)
  
  if ferr != nil {
    println(ferr)
  }

  return toHl7(xml.NewDecoder(file))
}
*/

//Pretty print in with indentation.
func PrintXmlDoc(doc Hl7Doc) {

  o, err := xml.MarshalIndent(doc, "  ", "    ")

  if err != nil {
    println(err)
  }
  
  os.Stdout.Write(o)
}

func XmlToString(doc Hl7Doc) string {
  o, err := xml.Marshal(doc)
  
  if err != nil{
    println(err)
  }

  return string(o)
}

func WriteXmlFile(doc Hl7Doc, filename string) {
  o, err := xml.Marshal(doc)

  if err != nil {
    println(err)
  }

	file, error := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0666)
	if(error != nil) {
		fmt.Printf("file error: %v\n", error)
	}
	
	file.Write(o)

}

//Loads an *xml.Document returned from DocFrom...
func EditNode(headerName string, childName string, doc *xmlx.Document, value string) *xmlx.Document {
  var newchild *xmlx.Node
  parent := doc.SelectNode("", headerName)
  
  child := parent.SelectNode("", childName)

  parent.RemoveChild(child)
  
  child.Value = value
  newchild = child

  parent.AddChild(newchild)

  return doc
}

//xmlx has some nice load functions. I wrap a few of them here.
//refer to xmlx source to see more of them.
func DocFromStream(r io.Reader) *xmlx.Document {
  doc := xmlx.New()
  doc.LoadStream(r, nil)

  return doc
}

func DocFromFile(filename string) *xmlx.Document{
  doc := xmlx.New()
  doc.LoadFile(filename, nil)

  return doc
}

//turn an xmlx node into an hl7 string
func NodeToHl7(node *xmlx.Node) string {
  var nodeNames []string

  var hl7slice []string = make([]string, len(node.Children) + 1)

  for i := 0; i < len(node.Children); i++ {
    nodeNames = append(nodeNames, (node.Children[i].Name).Local)

    nodeName := (node.Children[i].Name).Local
    nameParts := strings.Split(nodeName, ".")
    place, strconv_err := strconv.Atoi(nameParts[1])
    if strconv_err == nil {
      hl7slice[place] = node.Children[i].Value
    } else {
      fmt.Errorf("err: %v\n", strconv_err)
      hl7slice[i] = node.Children[i].Value
    }
  }

  return strings.Join(hl7slice, "|")
}

//turn a collection of xmlx nodes (*xmlx.Document) into an hl7 string.
func DocumentToHl7(doc *xmlx.Document) string {
  
  root := doc.Root
  root = root.Children[0]

  var hl7slice []string = make([]string, len(root.Children))

  for i := 0; i < len(root.Children); i++ {
    hl7slice[i] = concat(NodeToHl7(root.Children[i]), "\r")
  }
  return strings.Join(hl7slice, "\r")

}

//Concatenate two strings. 
//Go doesn't have a builtin way to do this. Sad.
func concat(string1 string, string2 string) string {
  var buff bytes.Buffer

  buff.WriteString(string1)
  buff.WriteString(string2)

  return buff.String()
}

func StringToXml(data string) Hl7Doc {
  return toXML(data)
}
