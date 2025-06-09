package unstructuredprocessor

import (
	"bytes"
	"encoding/json"
	"fmt"
	structuredoutput "llmdojo"
	"log"
	"os"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/ledongthuc/pdf"
	"github.com/openai/openai-go"
)

// ReadPDFContent reads and returns the content of a PDF file as human readable text.
func ReadPDFContent(pdfPath string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(pdfPath); err != nil {
		return "", fmt.Errorf("error accessing file: %v", err)
	}

	f, r, err := pdf.Open(pdfPath)
	if err != nil {
		return "", fmt.Errorf("error opening PDF: %v", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("error extracting text from PDF: %v", err)
	}

	buf.ReadFrom(b)
	return buf.String(), nil
}

type DocType string

const (
	RESUME       DocType = "RESUME"
	COVER_LETTER DocType = "COVER_LETTER"
	UNKNOWN      DocType = "UNKNOWN"
)

type DocDescriptor interface {
	DocType() DocType
}

func (ResumeFeatures) DocType() DocType {
	return RESUME
}

type CoverLetter struct{}

func (CoverLetter) DocType() DocType {
	return COVER_LETTER
}

func (c *CoverLetter) Features() (map[string]interface{}, error) {
	bytes, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}

	var features map[string]interface{}
	if err := json.Unmarshal(bytes, &features); err != nil {
		return nil, err
	}
	return features, nil
}

type DocClassification struct {
	DocType DocType `json:"docType" jsonschema:"description=The type of document that was classified (RESUME, COVER_LETTER, or UNKNOWN)"`
}

type ResumeFeatures struct {
	FirstName          string             `json:"firstName" jsonschema:"description=The first name of the candidate"`
	LastName           string             `json:"lastName" jsonschema:"description=The last name of the candidate"`
	Contact            Contact            `json:"contact" jsonschema:"description=Contact information of the candidate"`
	Education          Education          `json:"education" jsonschema:"description=The education details of the candidate"`
	YearsOfExperience  float32            `json:"yearsOfExperience" jsonschema:"description=The number of years of experience the candidate has"`
	Skills             []string           `json:"skills" jsonschema:"description=The skills possessed by the candidate"`
	WorkExperience     WorkExperience     `json:"workExperience" jsonschema:"description=The work experience of the candidate"`
	SalaryExpectation  float32            `json:"salaryExpectation" jsonschema:"description=The salary expectation of the candidate"`
	Location           string             `json:"location" jsonschema:"description=The location of the candidate"`
	OpenSourceProjects OpenSourceProjects `json:"openSourceProjects" jsonschema:"description=The open source projects the candidate has contributed to"`
}

type Contact struct {
	Email string `json:"email" jsonschema:"description=The email address of the candidate"`
	Phone string `json:"phone" jsonschema:"description=The phone number of the candidate"`
}
type Education []struct {
	Degree string `json:"degree" jsonschema:"description=The degree obtained by the candidate"`
}

type WorkExperience []struct {
	CompanyName string `json:"companyName" jsonschema:"description=The name of the company where the candidate worked"`
	Position    string `json:"position" jsonschema:"description=The position held by the candidate"`
}
type OpenSourceProjects []struct {
	ProjectName string `json:"projectName" jsonschema:"description=The name of the open source project the candidate contributed to"`
	GithubLink  string `json:"githubLink" jsonschema:"description=The GitHub link to the open source project"`
}

var docClassificationSchema = openai.ResponseFormatJSONSchemaJSONSchemaParam{
	Name:        "DocClassification",
	Description: openai.String("Classify the document into one of the following categories: Resume, Cover Letter, or Unknown."),
	Schema:      GenerateSchema[DocClassification](),
	Strict:      openai.Bool(true),
}

var ResumeFeaturesSchema = openai.ResponseFormatJSONSchemaJSONSchemaParam{
	Name:        "ResumeFeatures",
	Description: openai.String("Extract features from the resume."),
	Schema:      GenerateSchema[ResumeFeatures](),
	Strict:      openai.Bool(true),
}

func GenerateSchema[T any]() interface{} {

	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

// ClassifyDocument classifies the document content into one of the predefined categories.
// It uses the OpenAI API to generate a response based on the provided content.
// The function returns the classified document type or an error if the classification fails.
// The function takes a string parameter 'content' which contains the text of the document to be classified.
// It returns a DocType representing the classified document type and an error if any occurs during the classification process.
// The function uses a structured output format to define the expected response schema.
func ClassifyDocument(content string) (DocType, error) {

	conv := structuredoutput.NewChatContext(1)
	conv.AddMessage(openai.ChatCompletionMessageParamUnion{
		OfSystem: &openai.ChatCompletionSystemMessageParam{
			Content: openai.ChatCompletionSystemMessageParamContentUnion{
				OfString: openai.String("You are a document classification expert. Classify the document into one of the following categories: Resume, Cover Letter, or Unknown."),
			},
		},
	})

	conv.AddMessage(openai.ChatCompletionMessageParamUnion{
		OfUser: &openai.ChatCompletionUserMessageParam{
			Content: openai.ChatCompletionUserMessageParamContentUnion{
				OfString: openai.String(content),
			},
		},
	})

	agentResp, err := conv.GenerateResponseFromModel(docClassificationSchema)
	if err != nil {
		return "", fmt.Errorf("error generating response from model: %v", err)
	}

	var response structuredoutput.ModelResp
	if err := json.Unmarshal([]byte(agentResp), &response); err != nil {
		log.Printf("Error unmarshalling response: %v", err)
		return "", err
	}

	var docTypeResponse DocClassification
	if err := json.Unmarshal([]byte(response.Content), &docTypeResponse); err != nil {
		log.Printf("Error unmarshalling document type response: %v", err)
		return "", err
	}

	return docTypeResponse.DocType, nil
}

// ExtractDataFromResume extracts data from the resume content.
// It uses the OpenAI API to generate a response based on the provided content.
// The function returns a ResumeFeatures struct containing the extracted data or an error if the extraction fails.
// The function takes a string parameter 'content' which contains the text of the resume to be processed.
// It returns a pointer to a ResumeFeatures struct and an error if any occurs during the extraction process.
// The function uses a structured output format to define the expected response schema.
func ExtractDataFromResume(content string) (*ResumeFeatures, error) {

	conv := structuredoutput.NewChatContext(1)
	conv.AddMessage(openai.ChatCompletionMessageParamUnion{
		OfSystem: &openai.ChatCompletionSystemMessageParam{
			Content: openai.ChatCompletionSystemMessageParamContentUnion{
				OfString: openai.String("You are a resume data extraction expert. Extract the following information from the resume: contact information, education, years of experience, skills, and work experience."),
			},
		},
	})

	conv.AddMessage(openai.ChatCompletionMessageParamUnion{
		OfUser: &openai.ChatCompletionUserMessageParam{
			Content: openai.ChatCompletionUserMessageParamContentUnion{
				OfString: openai.String(content),
			},
		},
	})

	agentResp, err := conv.GenerateResponseFromModel(ResumeFeaturesSchema)
	if err != nil {
		return nil, fmt.Errorf("error generating response from model: %v", err)
	}
	var response structuredoutput.ModelResp
	if err := json.Unmarshal([]byte(agentResp), &response); err != nil {
		log.Printf("Error unmarshalling response: %v", err)
		return nil, err
	}
	var resumeData ResumeFeatures
	if err := json.Unmarshal([]byte(response.Content), &resumeData); err != nil {
		log.Printf("Error unmarshalling resume data response: %v", err)
		return nil, err
	}

	return &resumeData, nil
}

func ExtractFeatures(doc string) (DocType, DocDescriptor, error) {
	content, err := ReadPDFContent("../AD-Resume-v4.pdf")
	if err != nil {
		fmt.Println("Error:", err)
		return "", nil, err
	}
	// fmt.Println("PDF Content:", content)

	docType, err := ClassifyDocument(content)
	if err != nil {
		fmt.Println("Error:", err)
		return "", nil, err
	}
	fmt.Println("Document Type:", docType)

	switch strings.ToUpper(string(docType)) {
	case string(RESUME):
		resumeData, err := ExtractDataFromResume(content)
		if err != nil {
			fmt.Println("Error:", err)
			return "", nil, err
		}
		fmt.Printf("Resume Data: %+v\n", resumeData)
		return RESUME, resumeData, nil

	default:
		fmt.Print("This document type is not classified.")
		return UNKNOWN, nil, fmt.Errorf("unknown document type")
	}

}
