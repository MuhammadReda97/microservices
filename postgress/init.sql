-- Create the products table
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    category VARCHAR(100),
    brand VARCHAR(100),
    status VARCHAR(50),
    color VARCHAR(50)
    );

-- Create the cars table
CREATE TABLE IF NOT EXISTS cars (
    id SERIAL PRIMARY KEY,
    brand VARCHAR(50),
    model VARCHAR(50),
    model_year INT,
    color VARCHAR(20),
    max_speed INT,
    tire_size VARCHAR(20),
    weight INT,
    body VARCHAR(50),
    price INT
    );


COPY products(id, name, description, price, category, brand, status, color)
    FROM '/usr/share/postgres/products.csv'
    DELIMITER ','
    CSV HEADER;

COPY cars(id, brand, model, model_year, color, max_speed, tire_size, weight, body, price)
    FROM '/usr/share/postgres/cars.csv'
    DELIMITER ','
    CSV HEADER;
