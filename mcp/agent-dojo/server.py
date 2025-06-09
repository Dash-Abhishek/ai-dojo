# server.py
from fastmcp import FastMCP
from fastmcp.resources import TextResource
from openai import AzureOpenAI
import dotenv



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




mcpserver = FastMCP(name="MyMCP")
mcpserver.debug = True 


@mcpserver.tool(name="greet", description="Greet a user")
def greet(name: str) -> str:
    return f"Hello, {name}!"


@mcpserver.tool(name="GetServiceDetails", description="Given a path to a service, return the details")
def get_service_details(path: str) -> dict:
    return {
        "name": "ServiceX",
        "MACid": "1234567890",
        "Platform": "TSE",
        "Portfolio": "PortfolioX",
        "Criticality": "High",
        "owner": "John Doe",
        "email": "jhondoe@maersk.com",
        "description": "This is a sample service",
        "serviceHealthEndpoint": "https://example.com/health",
        "version": "1.0.0",
    }



@mcpserver.tool(name="TriggerHedwigAlert", description="Given a hedwig scope, trigger an alert")
def trigger_hedwig_alert(scope: str) -> dict:
    return {
        "status": "success",
        "message": f"Hedwig alert triggered for scope: {scope}",
    }
    


@mcpserver.tool(name="GetDependencies", description="Given a service name, return the dependencies")
def get_dependencies(svc_name: str) -> list:
    return ["svc1", "svc2", "svc3"]

if __name__ == "__main__":
    print(f"Starting server on port ")
    mcpserver.run(transport="sse")