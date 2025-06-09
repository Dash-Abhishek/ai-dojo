package main

import (
	"encoding/json"
	"fmt"
	structuredoutput "llmdojo"
	"log"

	"database/sql"

	"github.com/invopop/jsonschema"
	_ "github.com/mattn/go-sqlite3"
	"github.com/openai/openai-go"
)

type Step struct {
	Explanation string `json:"explanation"`
	// Output      string `json:"output"`
}
type AgentResponseFormat struct {
	Steps       []Step `json:"steps"`
	FinalOutput string `json:"finalOutput"`
}

var respSchema = openai.ResponseFormatJSONSchemaJSONSchemaParam{
	Name:        "SqlPipeline",
	Description: openai.String("SQL pipeline for generating SQL queries"),
	Schema:      GenerateSchema[AgentResponseFormat](),
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

const initialContext = `You are an expert in Databases SQLite, Python and data analysis.
			You need to are given this database schema and a a following question.
            You need to provide  correct SQL query to answer the question.
            You need to provide the SQL query only, & that has to be correct and without newline. 
            Do explain your reasoning in 1-3 steps and then finally provide the SQL query. 
            Database schema is attached below. Note the relationships between tables and the data types of each column.
            The database schema is as follows:
            
erDiagram
    orders || --o{
            order_items: contains
    orders || --o{
            order_payments: has
    orders || --o{
            order_reviews: has
    orders
        } | --|| customers : placed_by
    order_items } | --|| products : includes
    order_items }| --|| sellers : sold_by
sellers }| --|| geolocation : located_in
customers }| --|| geolocation : located_in

    orders {
        string order_id
        string customer_id
        string order_status
        datetime order_purchase_timestamp
        datetime order_approved_at
        datetime order_delivered_carrier_date
        datetime order_delivered_customer_date
        datetime order_estimated_delivery_date
}

    order_items {
        string order_id
        int order_item_id
        string product_id
        string seller_id
        datetime shipping_limit_date
        float price
        float freight_value
}

    order_payments {
        string order_id
        int payment_sequential
        string payment_type
        int payment_installments
        float payment_value
}

    order_reviews {
        string review_id
        string order_id
        int review_score
        string review_comment_title
        string review_comment_message
        datetime review_creation_date
        datetime review_answer_timestamp
}

    customers {
        string customer_id
        string customer_unique_id
        string customer_zip_code_prefix
        string customer_city
        string customer_state
}

    sellers {
        string seller_id
        string seller_zip_code_prefix
        string seller_city
        string seller_state
}

    products {
        string product_id
        string product_category_name
        int product_name_length
        int product_description_length
        int product_photos_qty
        float product_weight_g
        float product_length_cm
        float product_height_cm
        float product_width_cm
}

    geolocation {
        string geolocation_zip_code_prefix
        float geolocation_lat
        float geolocation_lng
        string geolocation_city
        string geolocation_state

}`

func main() {

	type Example struct {
		Question string
		Answer   string
	}
	examples := []Example{
		{
			Question: "Which seller has delivered the most orders to customers in Rio de Janeiro? [string: seller_id]",
			Answer:   "SELECT s.seller_id, COUNT(*) AS order_count FROM orders o JOIN customers c ON o.customer_id = c.customer_id JOIN sellers s ON o.seller_id = s.seller_id WHERE c.customer_city = 'rio de janeiro' AND o.order_status = 'delivered' GROUP BY s.seller_id ORDER BY order_count DESC LIMIT 1;",
		},
		{
			Question: "What's the average review score for 'beleza_saude' products?",
			Answer:   "SELECT AVG(r.review_score) AS avg_score FROM order_reviews r JOIN order_items oi ON r.order_id = oi.order_id JOIN products p ON oi.product_id = p.product_id WHERE p.product_category_name = 'beleza_saude';",
		},
	}

	testCases := []string{
		"Which seller has delivered the most orders to customers in Rio de Janeiro? [string: seller_id]",
		// "What's the average review score for products in the 'beleza_saude' category? [float: score]",
		// "How many sellers have completed orders worth more than 100,000 BRL in total? [integer: count]",
		// "Which product category has the highest rate of 5 - star reviews ? [string: category_name]",
		// "What's the most common payment installment count for orders over 1000 BRL? [integer: installments]",
		// "Which city has the highest average freight value per order? [string: city_name]",
		// "What's the most expensive product category based on average price? [string: category_name]",
		// "Which product category has the shortest average delivery time? [string: category_name]",
		// "How many unique customers have placed orders in the state of Sao Paulo? [integer: count]",
		// "What percentage of orders are delivered before the estimated delivery date ? [float: percentage]"
	}
	failedgenerations := 0
	for caseId, testCase := range testCases {

		conv := structuredoutput.NewChatContext(caseId)

		// Add system message to the conversation
		// This message is used to set the context for the conversation
		// Explains the role of the assistant, and introduces to the database schema
		// and the task at hand
		conv.AddMessage(openai.ChatCompletionMessageParamUnion{
			OfSystem: &openai.ChatCompletionSystemMessageParam{
				Content: openai.ChatCompletionSystemMessageParamContentUnion{
					OfString: openai.String(initialContext),
				},
			},
		})

		// few-short learning
		for _, example := range examples {
			conv.AddMessage(openai.ChatCompletionMessageParamUnion{
				OfUser: &openai.ChatCompletionUserMessageParam{
					Content: openai.ChatCompletionUserMessageParamContentUnion{
						OfString: openai.String(example.Question),
					},
				},
			})
			conv.AddMessage(openai.ChatCompletionMessageParamUnion{
				OfAssistant: &openai.ChatCompletionAssistantMessageParam{
					Content: openai.ChatCompletionAssistantMessageParamContentUnion{
						OfString: openai.String(example.Answer),
					},
				},
			})
		}

		// user question
		conv.AddMessage(openai.ChatCompletionMessageParamUnion{
			OfUser: &openai.ChatCompletionUserMessageParam{
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfString: openai.String(testCase),
				},
			},
		})

		// Uncomment this line to view the conversation
		// conv.ViewConversation()

		resp, err := conv.GenerateResponseFromModel(respSchema)
		if err != nil {
			log.Printf("Error generating response: %v", err)
			continue
		}
		var response structuredoutput.ModelResp
		if err := json.Unmarshal([]byte(resp), &response); err != nil {
			log.Printf("Error unmarshalling response: %v", err)
			continue
		}

		var agentResp AgentResponseFormat
		json.Unmarshal([]byte(response.Content), &agentResp)
		fmt.Printf("User query : %s\n", testCase)
		for i, step := range agentResp.Steps {
			fmt.Printf("Step %d:\n", i+1)
			fmt.Printf("Explanation:\n %s\n", step.Explanation)
		}
		fmt.Printf("Final Output:\n%s\n", agentResp.FinalOutput)
		// execute the SQL query
		results, err := ExecuteSQLQuery("/Users/adash/personal/ai-dojo/olist.sqlite", agentResp.FinalOutput)
		if err != nil {
			log.Printf("Error executing SQL query: %v", err)
			failedgenerations++
			continue
		}

		fmt.Printf("result: %+v\n", results)
		fmt.Println("--------------------------------------------------")

	}
	fmt.Printf("Failed generations: %d\n", failedgenerations)
	fmt.Printf("Total test cases: %d\n", len(testCases))
	fmt.Printf("Success rate: %.2f%%\n", (1-float64(failedgenerations)/float64(len(testCases)))*100)
	fmt.Println("--------------------------------------------------")
}

func ExecuteSQLQuery(dbPath string, query string) ([]map[string]interface{}, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]interface{}
	for rows.Next() {
		columnPointers := make([]interface{}, len(columns))
		columnValues := make([]interface{}, len(columns))
		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		rowMap := make(map[string]interface{})
		for i, colName := range columns {
			rowMap[colName] = columnValues[i]
		}
		results = append(results, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}
