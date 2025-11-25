-- Help Desk Service Database Schema
-- Migration: 001_initial_schema.sql
-- Note: All foreign keys to User/Equipment removed, using integer references instead

-- Create help_desk_type table
CREATE TABLE IF NOT EXISTS "Help_desk_type" (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    team_id INTEGER,
    is_active BOOLEAN,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create team table (help desk teams)
CREATE TABLE IF NOT EXISTS "Team" (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create help_desk table (removed foreign key to User, keeping portal_user_id as integer)
CREATE TABLE IF NOT EXISTS "Help_desk" (
    id SERIAL PRIMARY KEY,
    name CHAR NOT NULL,
    create_date DATE NOT NULL,
    help_desk_type_id INTEGER NOT NULL,
    portal_user_id INTEGER NOT NULL, -- Reference to user in Auth Service (no foreign key)
    description TEXT NOT NULL,
    state CHAR NOT NULL,
    resolve_date DATE NOT NULL,
    help_desk_id INTEGER, -- Parent ticket ID (self-reference)
    project_id INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (help_desk_type_id) REFERENCES "Help_desk_type"(id),
    FOREIGN KEY (help_desk_id) REFERENCES "Help_desk"(id)
);

CREATE INDEX idx_help_desk_portal_user_id ON "Help_desk"(portal_user_id);
CREATE INDEX idx_help_desk_state ON "Help_desk"(state);
CREATE INDEX idx_help_desk_type_id ON "Help_desk"(help_desk_type_id);

-- Create ticket_comments table
CREATE TABLE IF NOT EXISTS ticket_comments (
    id SERIAL PRIMARY KEY,
    ticket_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL, -- Reference to user in Auth Service (no foreign key)
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (ticket_id) REFERENCES "Help_desk"(id) ON DELETE CASCADE
);

CREATE INDEX idx_ticket_comments_ticket_id ON ticket_comments(ticket_id);
CREATE INDEX idx_ticket_comments_user_id ON ticket_comments(user_id);

-- Create documents table (for ticket attachments, removed foreign keys to Equipment/Supplier)
CREATE TABLE IF NOT EXISTS "Documents" (
    id SERIAL PRIMARY KEY,
    name CHAR NOT NULL,
    document_size CHAR NOT NULL,
    picture BYTEA NOT NULL,
    document_type CHAR NOT NULL,
    equipment_id INTEGER DEFAULT 0, -- Reference only (no foreign key)
    supplier_id INTEGER DEFAULT 0, -- Reference only (no foreign key)
    help_desk_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (help_desk_id) REFERENCES "Help_desk"(id) ON DELETE CASCADE
);

CREATE INDEX idx_documents_help_desk_id ON "Documents"(help_desk_id);

-- Create ticket_attachments table (links documents to tickets)
CREATE TABLE IF NOT EXISTS ticket_attachments (
    id SERIAL PRIMARY KEY,
    ticket_id INTEGER NOT NULL,
    document_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (ticket_id) REFERENCES "Help_desk"(id) ON DELETE CASCADE,
    FOREIGN KEY (document_id) REFERENCES "Documents"(id) ON DELETE CASCADE
);

CREATE INDEX idx_ticket_attachments_ticket_id ON ticket_attachments(ticket_id);

-- Create user_help_desk join table (removed foreign key to User)
CREATE TABLE IF NOT EXISTS user_help_desk (
    user_id INTEGER NOT NULL, -- Reference to user in Auth Service (no foreign key)
    help_desk_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, help_desk_id),
    FOREIGN KEY (help_desk_id) REFERENCES "Help_desk"(id) ON DELETE CASCADE
);

CREATE INDEX idx_user_help_desk_user_id ON user_help_desk(user_id);
CREATE INDEX idx_user_help_desk_help_desk_id ON user_help_desk(help_desk_id);

-- Create equipment_help_desk join table (removed foreign key to Equipment)
CREATE TABLE IF NOT EXISTS equipment_help_desk (
    equipment_id INTEGER NOT NULL, -- Reference only (no foreign key)
    help_desk_id INTEGER NOT NULL,
    PRIMARY KEY (equipment_id, help_desk_id),
    FOREIGN KEY (help_desk_id) REFERENCES "Help_desk"(id) ON DELETE CASCADE
);

-- Create team_members table (removed foreign key to User)
CREATE TABLE IF NOT EXISTS team_members (
    user_id INTEGER NOT NULL, -- Reference to user in Auth Service (no foreign key)
    help_desk_team_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, help_desk_team_id),
    FOREIGN KEY (help_desk_team_id) REFERENCES "Team"(id) ON DELETE CASCADE
);

-- Create transaction_type table
CREATE TABLE IF NOT EXISTS "Transaction_type" (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT
);

-- Create Help_desk_transaction table
CREATE TABLE IF NOT EXISTS "Help_desk_transaction" (
    id SERIAL PRIMARY KEY,
    name CHAR NOT NULL,
    transaction_type_id INTEGER NOT NULL,
    date_time DATE NOT NULL,
    "time" FLOAT NOT NULL,
    help_desk_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (transaction_type_id) REFERENCES "Transaction_type"(id),
    FOREIGN KEY (help_desk_id) REFERENCES "Help_desk"(id) ON DELETE CASCADE
);

CREATE INDEX idx_help_desk_transaction_help_desk_id ON "Help_desk_transaction"(help_desk_id);

-- Create Help_desk_transaction_user join table (removed foreign key to User)
CREATE TABLE IF NOT EXISTS "Help_desk_transaction_user" (
    transaction_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL, -- Reference to user in Auth Service (no foreign key)
    PRIMARY KEY (transaction_id, user_id),
    FOREIGN KEY (transaction_id) REFERENCES "Help_desk_transaction"(id) ON DELETE CASCADE
);

-- Create Help_desk_rating table
CREATE TABLE IF NOT EXISTS "Help_desk_rating" (
    id SERIAL PRIMARY KEY,
    help_desk_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL, -- Reference to user in Auth Service (no foreign key)
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (help_desk_id) REFERENCES "Help_desk"(id) ON DELETE CASCADE
);

CREATE INDEX idx_help_desk_rating_help_desk_id ON "Help_desk_rating"(help_desk_id);

