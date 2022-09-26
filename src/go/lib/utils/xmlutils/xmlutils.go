package xmlutils

import (
	"github.com/beevik/etree"
	"github.com/ma314smith/signedxml"

	"github.com/pkg/errors"

	"github.com/myfintech/ark/src/go/lib/log"
)

// XMLDigitalSignatureC14NTransform will perform a canonical transform on an XML string and
// return an XMLDocument object. See these links for more details
// https://www.w3.org/TR/xml-exc-c14n/
// https://www.w3.org/TR/xml-exc-c14n/#ref-XML-C14N
func XMLDigitalSignatureC14NTransform(xmlText string, stripComments bool, stripWhitespace bool) (*etree.Document, error) {
	canonicalizer := signedxml.ExclusiveCanonicalization{
		WithComments: stripComments == false,
	}

	transformedXML, err := canonicalizer.Process(xmlText, "")
	if err != nil {
		return nil, errors.Wrap(err, "failed to canonicalize XML")
	}

	doc, err := XMLDocument(transformedXML)
	if err != nil {
		return doc, err
	}

	if stripWhitespace {
		StripWhitespace(doc)
	}

	return doc, nil
}

func StripWhitespace(document *etree.Document) {
	ProcessNodesRecursive(document.Root(), func(token etree.Token) {
		switch c := token.(type) {
		case *etree.CharData:
			if IsWhitespace(c.Data) {
				c.Data = ""
			}
		}
	})
}

type TokenProcessor func(node etree.Token)

func ProcessNodesRecursive(node *etree.Element, processToken TokenProcessor) {
	for _, child := range node.Child {
		switch child := child.(type) {
		case *etree.Element:
			processToken(child)
			ProcessNodesRecursive(child, processToken)
		default:
			processToken(child)
		}
	}
}

// IsWhitespace returns true if the byte slice contains only
// whitespace characters.
func IsWhitespace(s string) bool {
	for i := 0; i < len(s); i++ {
		if c := s[i]; c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			return false
		}
	}
	return true
}

// XMLDocument parses an XML string and returns a virtual DOM
func XMLDocument(xml string) (*etree.Document, error) {
	doc := etree.NewDocument()
	doc.WriteSettings = etree.WriteSettings{
		CanonicalEndTags: true,
		CanonicalText:    true,
		CanonicalAttrVal: true,
		UseCRLF:          false,
	}
	// FIXME: think about this... make optional?
	// more on this... it has bit me more than once... maybe don't
	// doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)
	if err := doc.ReadFromString(xml); err != nil {
		return doc, errors.Wrapf(err, "failed to parse XML into tree %s", xml)
	}
	return doc, nil
}

// PrettyPrint returns a formatted XML document as a string

type XMLDigestSignature struct {
	Doc                                   *etree.Document
	MessageDigestElementPath              string
	MessageDigestSignatureKey             string
	MessageDigestSignatureEncoding        string
	MessageDigestSignatureRootElementPath string
	MessageDigestSignatureAlgorithm       string
	Pretty                                bool
}

// XMLDigestSigner a callback function to generate a cryptographic signature given a canonical document
type XMLDigestSigner func(tree *etree.Document, document, signatureAlg, signatureKey, encodingFormat string) (string, error)

// Inject executes a callback which receives the canonicalXML for which a signature should be generated and expects a returned string
func (d *XMLDigestSignature) Inject(generateSignature XMLDigestSigner) error {
	digestElem := d.Doc.FindElement(d.MessageDigestElementPath)

	if digestElem == nil {
		return errors.Errorf("Cannot locate digest element with %s", d.MessageDigestElementPath)
	}
	log.Debugf("Digest element located at %s", digestElem.GetPath())

	digestElemParent := digestElem.Parent()
	log.Debugf("Digest element parent located at %s", digestElemParent.GetPath())

	if digestElemParent.RemoveChild(digestElem) == nil {
		return errors.New("Cannot remove digest element at path %s because it has no parent")
	}

	if d.Pretty {
		d.Doc.Indent(4)
	}

	rootSignatureElem := d.Doc.FindElement(d.MessageDigestSignatureRootElementPath)

	if rootSignatureElem == nil {
		return errors.Errorf("failed to locate messageDigestSignatureRootElement %s", d.MessageDigestSignatureRootElementPath)
	}

	log.Debugf("Digest signature root element located at %s", rootSignatureElem.GetPath())

	rootDocument := etree.NewDocument()
	rootDocument.SetRoot(rootSignatureElem.Copy())
	docWithoutDigestElem, err := rootDocument.WriteToString()
	if err != nil {
		return errors.Wrap(err, "failed to output document before calculating signature")
	}

	canonicalXML, err := XMLDigitalSignatureC14NTransform(docWithoutDigestElem, true, true)
	if err != nil {
		return errors.Wrap(err, "failed to build canonical XML")
	}

	canonicalXMLString, err := canonicalXML.WriteToString()
	if err != nil {
		return errors.Wrap(err, "failed to write canonical xml to string")
	}

	signature, err := generateSignature(canonicalXML, canonicalXMLString, d.MessageDigestSignatureAlgorithm, d.MessageDigestSignatureKey, d.MessageDigestSignatureEncoding)

	if err != nil {
		return errors.Wrap(err, "failed to calculate signature")
	}

	digestElemParent.AddChild(digestElem)
	digestElem.SetText(signature)

	return nil
}
