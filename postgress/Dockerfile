FROM postgres:latest

# Set environment variables
ENV POSTGRES_USER=afghany
ENV POSTGRES_PASSWORD=mypassword
ENV POSTGRES_DB=mydatabase

# Create the directory inside the container
RUN mkdir -p /usr/share/postgres/

# Copy the CSV files into the container
COPY cars.csv /usr/share/postgres/cars.csv
COPY products.csv /usr/share/postgres/products.csv

# Copy the initialization script
COPY init.sql /docker-entrypoint-initdb.d/

# Set correct permissions
RUN chmod 644 /usr/share/postgres/*.csv
RUN chmod 644 /docker-entrypoint-initdb.d/init.sql

# Expose PostgreSQL port
EXPOSE 5432
