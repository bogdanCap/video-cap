CREATE USER db_user WITH PASSWORD 'secure_password_123';
CREATE DATABASE app_record OWNER db_user;
GRANT ALL PRIVILEGES ON DATABASE app_record TO db_user;