import ollama from 'ollama'
import sqlite3 from 'sqlite3'



const initialPrompt = [
    {
        role: 'system',
        content: `You are an expert in Databases SQLite, Python and data analysis.
            You need to are given this database schema and a a following question.
           
            
            You need to provide  correct SQL query to answer the question.
            You need to provide the SQL query only, & that has to be correct and without newline. 
            Do not provide any explanation or any other text.
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

}`},
]

async function generateSQLQuery(question) {
    const response = await ollama.chat({
        model: 'llama3.2',
        messages: [...initialPrompt, { role: 'user', content: question }],
        tools: [],
    })
    return response.message.content
}

async function ExecuteSQLQuery(query) {

    return new Promise((resolve, reject) => {

        const db = new sqlite3.Database('/Users/adash/personal/ai-dojo/olist.sqlite', (err) => {
            if (err) {
                reject(err);
            }
        });
        db.all(query, [], function (err, rows) {
            if (err) {
                reject(err);
            }
            resolve(rows);
        });

    })



}

async function testSqlQuery() {
    var userQuestions = [
        "Which seller has delivered the most orders to customers in Rio de Janeiro? [string: seller_id]",
        "What's the average review score for products in the 'beleza_saude' category? [float: score]",
        "How many sellers have completed orders worth more than 100,000 BRL in total? [integer: count]",
        "Which product category has the highest rate of 5 - star reviews ? [string: category_name]",
        "What's the most common payment installment count for orders over 1000 BRL? [integer: installments]",
        "Which city has the highest average freight value per order? [string: city_name]",
        "What's the most expensive product category based on average price? [string: category_name]",
        "Which product category has the shortest average delivery time? [string: category_name]",
        "How many unique customers have placed orders in the state of Sao Paulo? [integer: count]",
        "What percentage of orders are delivered before the estimated delivery date ? [float: percentage]"
    ]


    for (let i = 0; i < userQuestions.length; i++) {
        const question = userQuestions[i]
        let query = await generateSQLQuery(question)
        console.log(`Question: ${question}`)
        console.log(`SQL Query: ${query}`)
        try {
            let rows = await ExecuteSQLQuery(query)
            console.log(`Query executed successfully:`, rows);
        } catch (err) {
            console.error(`Error executing query: ${err.message}`);
        }

        console.log('--------------------------------------------------------')


    }
}

testSqlQuery()
