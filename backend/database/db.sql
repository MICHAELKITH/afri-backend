CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(150) NOT NULL UNIQUE,
    phone_number VARCHAR(20) UNIQUE,
    country VARCHAR(100) NOT NULL,
    study_level VARCHAR(100) NOT NULL,
    field_of_study VARCHAR(150) NOT NULL,
    learning_goals TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
