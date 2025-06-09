import os
from openai import AzureOpenAI
import dotenv
from agents import Agent, Runner
from agents.mcp import MCPServerSse
import PyPDF2
from pydantic import BaseModel, Field


# Load environment variables from .env file
dotenv.load_dotenv()
# Configure Azure OpenAI credentials
azure_endpoint = os.getenv("AZURE_OPENAI_ENDPOINT")
openai_api_key = os.getenv("OPENAI_API_KEY")
azure_openai_api_version = os.getenv("AZURE_OPENAI_API_VERSION")
azure_openai_model = os.getenv("AZURE_OPENAI_DEPLOYMENT")

client = AzureOpenAI(
    azure_endpoint=azure_endpoint,
    api_key=openai_api_key,
    api_version=azure_openai_api_version
)

# # Set up MCP server client pointing to our FastAPI server
tools_server = MCPServerSse(name="MyMCP", params={"url": "http://localhost:8000/sse"})


class WorkHistory(BaseModel):
    Company: str = Field(..., description="Company name")
    Role: str = Field(..., description="Role at the company")
    Duration: str = Field(..., description="Duration of employment")

class CandidateFeatures(BaseModel):
    name: str = Field(..., description="Name of the candidate")
    yearsOfExp: str = Field(..., description="Years of experience")
    skills: list[str] = Field(..., description="List of skills")
    location: str = Field(..., description="Location of the candidate")
    workHistory: list[WorkHistory] = Field(..., description="Work history of the candidate")

class AgentHooks:
    async def on_start(self, context, agent):
        print(f"Agent {agent.name} started with context: {context}")

    async def on_input(self, context, input):
        print(f"Input to the agent: {input}")

    async def on_output(self, context, output):
        print(f"Output from the agent: {output}")

    async def on_error(self, context, error):
        print(f"Error in the agent: {error}")
    
    async def on_end(self, context, output, error):
        print(f"Agent ended with output:")

    async def on_agent_end(self, context, output):
        print(f"Agent {context.agent.name} ended with output: {output}")
    

# onboarding_agent = Agent(
#     name="Candidate Onboarding Agent",
#     model="gpt-4",  # might be configured internally to use Azure via openai library
#     instructions="You are a an onboarding agent, you need to onboard candidates based on the onboarding checklist. You have access to the mcp server, please use it to onboard the candidate.",
#     handoff_description="An agent that can onboard candidates based on the onboarding checklist.",
#     mcp_servers=[tools_server],
#     # onboarding_agents=[screening_agent],
# )





resume_analyzer_agent = Agent(
    name="Candidate Screening Agent",
    model="gpt-4o",  # might be configured internally to use Azure via openai library
    instructions=(
        "You are a candidate screening agent. You will receive a candidate's resume as input context. "
        "Your task is to analyze the resume and extract the following details: "
        "name, years of experience, skills, location, and work history. "
        "Always base your response on the provided resume text."
        "Do not make up any information."
    ),
    handoff_description="An agent that can screen candidates resume.",
    input_type=str,
    # mcp_servers=[tools_server],
    output_type=CandidateFeatures,
    hooks=AgentHooks(),
)

hr_agent = Agent(
    name="Hr Agent",
    model="gpt-4o",  # might be configured internally to use Azure via openai library
    instructions=("You are an HR agent responsible for managing candidate screening and onboarding. "
        "You will receive a candidate's resume and must use helper agents to analyze it. "
        "Ensure the resume is passed to the appropriate agent for processing"
        ),
    # mcp_servers=[tools_server],
    handoffs=[resume_analyzer_agent],
        
    output_type=CandidateFeatures,
)

# assistant = client.beta.assistants.create(
#         name="Candidate Feature Extraction",
#         description="Extract candidate features from resume",
#         messages=[
#             {
#                 "role": "user",
#                 "content": f"Extract candidate features from the following resume: {resume}",
#             }
#         ],
#         response_format = CandidateFeatures,
#         temperature=0
#     )

def extract_text_from_pdf(pdf_path: str) -> str:
    """
    Extracts all text from a PDF file.

    Args:
        pdf_path (str): The path to the PDF file.

    Returns:
        str: The extracted text from the PDF.
    """
    extracted_text = ""
    try:
        with open(pdf_path, "rb") as pdf_file:
            reader = PyPDF2.PdfReader(pdf_file)
            for page in reader.pages:
                extracted_text += page.extract_text()
    except Exception as e:
        print(f"Error reading PDF file: {e}")
    return extracted_text





def main():

    resume_path = "./resume.pdf"

    # Extract text from the PDF resume
    resume_text = extract_text_from_pdf(resume_path)
    if not resume_text:
        print("No text extracted from the PDF.")
        os._exit(1)

    # print("Extracted Resume Text:", resume_text)
    # Connect to the MCP server first
    # await tools_server.connect()
    # Run the agent with a sample input
    # try:
        # Run the agent with a sample input
    result = Runner.run_sync(
        starting_agent=hr_agent, 
        context=resume_text,
        input="process candidates resume and extract features",
        )
    print(result)
    # except Exception as e:
    #     print(f"Error: {e}")
    # finally:
    #     # Cleanup: disconnect the connection when done
    #     try:
            # tools_server.cleanup()
        # except Exception as cleanup_error:
            # print(f"Cleanup error: {cleanup_error}")
    

if __name__ == "__main__":
   main()
