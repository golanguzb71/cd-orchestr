CREATE TABLE IF NOT EXISTS server_connections
(
    ip_address varchar NOT NULL,
    username   varchar NOT NULL,
    password   varchar NOT NULL,
    port       int     NOT NULL,
    created_at timestamp DEFAULT now()
);
