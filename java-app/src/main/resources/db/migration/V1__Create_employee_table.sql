-- Create employees table
CREATE TABLE employees (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    phone VARCHAR(20),
    department VARCHAR(100),
    position VARCHAR(100),
    hire_date DATE NOT NULL,
    salary DECIMAL(10,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_email (email),
    INDEX idx_department (department),
    INDEX idx_position (position),
    INDEX idx_hire_date (hire_date)
);

-- Insert sample data
INSERT INTO employees (first_name, last_name, email, phone, department, position, hire_date, salary) VALUES
('John', 'Doe', 'john.doe@company.com', '+1-555-0101', 'Engineering', 'Software Engineer', '2023-01-15', 75000.00),
('Jane', 'Smith', 'jane.smith@company.com', '+1-555-0102', 'Engineering', 'Senior Software Engineer', '2022-03-10', 95000.00),
('Michael', 'Johnson', 'michael.johnson@company.com', '+1-555-0103', 'Marketing', 'Marketing Manager', '2023-02-20', 65000.00),
('Emily', 'Davis', 'emily.davis@company.com', '+1-555-0104', 'HR', 'HR Specialist', '2023-04-05', 55000.00),
('David', 'Wilson', 'david.wilson@company.com', '+1-555-0105', 'Engineering', 'DevOps Engineer', '2022-11-12', 80000.00),
('Sarah', 'Brown', 'sarah.brown@company.com', '+1-555-0106', 'Finance', 'Financial Analyst', '2023-06-01', 60000.00);
