-- scripts/init.sql
CREATE DATABASE IF NOT EXISTS short_url CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE short_url;

CREATE TABLE IF NOT EXISTS links (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    short_code VARCHAR(10) NOT NULL UNIQUE,
    long_url TEXT NOT NULL,
    expires_at TIMESTAMP NULL,
    click_count BIGINT UNSIGNED DEFAULT 0,
    status ENUM('active', 'disabled', 'expired') DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    updated_by VARCHAR(100),
    description VARCHAR(100),
    delete_flag varchar(1),
    version INT UNSIGNED DEFAULT 0,
    INDEX idx_short_code (short_code),
    INDEX idx_created_by (created_by),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
    );

CREATE TABLE IF NOT EXISTS click_stats (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    short_code VARCHAR(10) NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    referer TEXT,
    country VARCHAR(2),
    region VARCHAR(100),
    city VARCHAR(100),
    device_type VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    updated_by VARCHAR(100),
    description VARCHAR(100),
    delete_flag varchar(1),
    version INT UNSIGNED DEFAULT 0,
    INDEX idx_short_code (short_code),
    INDEX idx_created_at (created_at)
    );