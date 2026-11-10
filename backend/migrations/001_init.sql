-- Idempotent helpdesk schema. Safe to run more than once.
-- users.updated_at is kept because the user repository reads and writes it.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE users ADD COLUMN IF NOT EXISTS name TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS email TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS role TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'users_role_check') THEN
        ALTER TABLE users
            ADD CONSTRAINT users_role_check
            CHECK (role IN ('requester', 'agent', 'admin'));
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_index i
        JOIN pg_attribute a
          ON a.attrelid = i.indrelid
         AND a.attnum = ANY (i.indkey)
        WHERE i.indrelid = 'public.users'::regclass
          AND i.indisunique
          AND i.indnkeyatts = 1
          AND a.attname = 'email'
    ) THEN
        CREATE UNIQUE INDEX idx_users_email ON users (email);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_users_role ON users (role);

CREATE TABLE IF NOT EXISTS tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    category TEXT NOT NULL,
    priority TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'open',
    requester_id UUID NOT NULL,
    assignee_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE tickets ADD COLUMN IF NOT EXISTS title TEXT;
ALTER TABLE tickets ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE tickets ADD COLUMN IF NOT EXISTS category TEXT;
ALTER TABLE tickets ADD COLUMN IF NOT EXISTS priority TEXT;
ALTER TABLE tickets ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'open';
ALTER TABLE tickets ADD COLUMN IF NOT EXISTS requester_id UUID;
ALTER TABLE tickets ADD COLUMN IF NOT EXISTS assignee_id UUID;
ALTER TABLE tickets ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE tickets ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'tickets_priority_check') THEN
        ALTER TABLE tickets
            ADD CONSTRAINT tickets_priority_check
            CHECK (priority IN ('low', 'medium', 'high', 'urgent'));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'tickets_status_check') THEN
        ALTER TABLE tickets
            ADD CONSTRAINT tickets_status_check
            CHECK (status IN ('open', 'in_progress', 'resolved', 'closed'));
    END IF;
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint c
        JOIN pg_attribute a
          ON a.attrelid = c.conrelid
         AND a.attnum = ANY (c.conkey)
        WHERE c.conrelid = 'public.tickets'::regclass
          AND c.contype = 'f'
          AND array_length(c.conkey, 1) = 1
          AND a.attname = 'requester_id'
    ) THEN
        ALTER TABLE tickets
            ADD CONSTRAINT tickets_requester_id_fkey
            FOREIGN KEY (requester_id) REFERENCES users (id) ON DELETE RESTRICT;
    END IF;
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint c
        JOIN pg_attribute a
          ON a.attrelid = c.conrelid
         AND a.attnum = ANY (c.conkey)
        WHERE c.conrelid = 'public.tickets'::regclass
          AND c.contype = 'f'
          AND array_length(c.conkey, 1) = 1
          AND a.attname = 'assignee_id'
    ) THEN
        ALTER TABLE tickets
            ADD CONSTRAINT tickets_assignee_id_fkey
            FOREIGN KEY (assignee_id) REFERENCES users (id) ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets (status);
CREATE INDEX IF NOT EXISTS idx_tickets_priority ON tickets (priority);
CREATE INDEX IF NOT EXISTS idx_tickets_requester_id ON tickets (requester_id);
CREATE INDEX IF NOT EXISTS idx_tickets_assignee_id ON tickets (assignee_id);
CREATE INDEX IF NOT EXISTS idx_tickets_created_at ON tickets (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tickets_requester_created ON tickets (requester_id, created_at DESC);

CREATE TABLE IF NOT EXISTS comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL,
    author_id UUID NOT NULL,
    body TEXT NOT NULL,
    is_internal BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE comments ADD COLUMN IF NOT EXISTS ticket_id UUID;
ALTER TABLE comments ADD COLUMN IF NOT EXISTS author_id UUID;
ALTER TABLE comments ADD COLUMN IF NOT EXISTS body TEXT;
ALTER TABLE comments ADD COLUMN IF NOT EXISTS is_internal BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE comments ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint c
        JOIN pg_attribute a
          ON a.attrelid = c.conrelid
         AND a.attnum = ANY (c.conkey)
        WHERE c.conrelid = 'public.comments'::regclass
          AND c.contype = 'f'
          AND array_length(c.conkey, 1) = 1
          AND a.attname = 'ticket_id'
    ) THEN
        ALTER TABLE comments
            ADD CONSTRAINT comments_ticket_id_fkey
            FOREIGN KEY (ticket_id) REFERENCES tickets (id) ON DELETE CASCADE;
    END IF;
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint c
        JOIN pg_attribute a
          ON a.attrelid = c.conrelid
         AND a.attnum = ANY (c.conkey)
        WHERE c.conrelid = 'public.comments'::regclass
          AND c.contype = 'f'
          AND array_length(c.conkey, 1) = 1
          AND a.attname = 'author_id'
    ) THEN
        ALTER TABLE comments
            ADD CONSTRAINT comments_author_id_fkey
            FOREIGN KEY (author_id) REFERENCES users (id) ON DELETE RESTRICT;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_comments_ticket_id ON comments (ticket_id);
CREATE INDEX IF NOT EXISTS idx_comments_author_id ON comments (author_id);
CREATE INDEX IF NOT EXISTS idx_comments_ticket_created ON comments (ticket_id, created_at);
