-- Idempotent seed. Password for every account is password123 (bcrypt).
-- Re-running does nothing when the email already exists.

INSERT INTO users (id, name, email, password_hash, role)
VALUES
    (
        '11111111-1111-4111-8111-111111111111',
        'Requester',
        'requester@example.com',
        '$2a$10$O92k1INcVHhw8WsW0zui8.X5sMdUj0MlNATiZ1ySQhwQpw2PbFe.q',
        'requester'
    ),
    (
        '22222222-2222-4222-8222-222222222222',
        'Agent',
        'agent@example.com',
        '$2a$10$O92k1INcVHhw8WsW0zui8.X5sMdUj0MlNATiZ1ySQhwQpw2PbFe.q',
        'agent'
    ),
    (
        '33333333-3333-4333-8333-333333333333',
        'Admin',
        'admin@example.com',
        '$2a$10$O92k1INcVHhw8WsW0zui8.X5sMdUj0MlNATiZ1ySQhwQpw2PbFe.q',
        'admin'
    )
ON CONFLICT (email) DO NOTHING;
