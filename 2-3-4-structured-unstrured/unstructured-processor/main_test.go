package unstructuredprocessor

import (
	"fmt"
	"math"
	"slices"
	"testing"
)

var resumeEvals = []struct {
	Resume         string
	ActualFeatures ResumeFeatures
	AccuracyScore  int
}{
	{"../AD-Resume-v4.pdf",
		ResumeFeatures{
			FirstName: "Abhishek",
			LastName:  "Dash",
			Contact: Contact{
				Email: "abhishekmicro@hotmail.com",
				Phone: "+918093958952",
			},
			Skills:    []string{"Go", "Node.js", "JavaScript", "HTML/CSS", "React.js", "Langchain", "Microservices", "API", "REST", "GRPC", "Cloud Computing", "System Design", "GitOpps", "API Gateway", "Temporal", "Postgres", "OpenAI", "LLM", "Docker", "Kubernetes", "Azure"},
			Education: Education{{Degree: "Bachelor of Technology (B.Tech.), Electrical and Electronics Engineering"}},
			WorkExperience: WorkExperience{
				{CompanyName: "A.P. Moller - Maersk", Position: "Senior Software Engineer"},
				{CompanyName: "Lowe's", Position: "Senior Software Engineer"},
				{CompanyName: "Hexaware Technologies", Position: "Research & Development Engineer"},
			},
			YearsOfExperience: 8.0,
		},
		8,
	},
	{
		"../AryanResume.pdf",
		ResumeFeatures{
			FirstName: "Aryan",
			LastName:  "Dash",
			Contact: Contact{
				Email: "aryanbgr20@gmail.com",
				Phone: "+919078302716",
			},
			Education:         Education{{Degree: "B.Tech in Electronics and Tele-Communication Engineering"}},
			Skills:            []string{"Python", "SQL", "PySpark", "Databricks", "Azure Data Factory", "Azure", "SQL Database", "Azure Synapse Analytics", "Azure Data Lake Storage", "MS SQL Server", "Data Warehousing", "ETL Processes"},
			YearsOfExperience: 2.0,
			WorkExperience: WorkExperience{
				{CompanyName: "SymphonyAI", Position: "Associate Software Engineer"},
			},
		},
		8,
	},
}

func TestResumeFeatureExtraction(t *testing.T) {

	for id, eval := range resumeEvals {
		content, err := ReadPDFContent(eval.Resume)
		if err != nil {
			t.Fatalf("Error reading PDF content: %v", err)
		}

		resumeData, err := ExtractDataFromResume(content)
		if err != nil {
			t.Fatalf("Error extracting data from resume: %v", err)
		}

		if resumeData == nil {
			t.Fatal("Extracted resume data is nil")
		}

		// Compare the extracted data with the expected data
		if resumeData.FirstName != eval.ActualFeatures.FirstName {
			t.Errorf("Expected FirstName: %s, got: %s", eval.ActualFeatures.FirstName, resumeData.FirstName)
			eval.AccuracyScore--
		}
		if resumeData.LastName != eval.ActualFeatures.LastName {
			t.Errorf("Expected LastName: %s, got: %s", eval.ActualFeatures.LastName, resumeData.LastName)
			eval.AccuracyScore--
		}
		if resumeData.Contact.Email != eval.ActualFeatures.Contact.Email {
			t.Errorf("Expected Email: %s, got: %s", eval.ActualFeatures.Contact.Email, resumeData.Contact.Email)
			eval.AccuracyScore--
		}
		if resumeData.Contact.Phone != eval.ActualFeatures.Contact.Phone {
			t.Errorf("Expected Phone: %s, got: %s", eval.ActualFeatures.Contact.Phone, resumeData.Contact.Phone)
			eval.AccuracyScore--
		}

		if len(resumeData.Skills) != len(eval.ActualFeatures.Skills) {
			t.Errorf("Expected Skills: %v, got: %v", eval.ActualFeatures.Skills, resumeData.Skills)
			eval.AccuracyScore--
		}
		if len(resumeData.Education) != len(eval.ActualFeatures.Education) {
			t.Errorf("Expected Education: %v, got: %v", eval.ActualFeatures.Education, resumeData.Education)
			eval.AccuracyScore--
		}
		if len(resumeData.WorkExperience) == len(eval.ActualFeatures.WorkExperience) {
			for _, work := range resumeData.WorkExperience {
				if !slices.Contains(eval.ActualFeatures.WorkExperience, work) {
					t.Errorf("Expected WorkExperience: %v, got: %v", eval.ActualFeatures.WorkExperience, work)
				}
			}
			eval.AccuracyScore--
		} else {
			t.Error("Expected WorkExperience length does not match")
			eval.AccuracyScore--
		}
		if math.Abs(float64(resumeData.YearsOfExperience-eval.ActualFeatures.YearsOfExperience)) > 0.01 {
			t.Errorf("Expected YearsOfExperience: %f, got: %f", eval.ActualFeatures.YearsOfExperience, resumeData.YearsOfExperience)
			eval.AccuracyScore--
		}

		fmt.Printf("eval %d Accuracy Score: %d/8\n", id, eval.AccuracyScore)
	}

}
