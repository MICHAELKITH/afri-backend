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
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'student' NOT NULL,
    password_reset_token VARCHAR(255),
    password_reset_expires TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);