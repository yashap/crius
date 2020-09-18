CREATE TABLE IF NOT EXISTS service (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(65535) UNIQUE NOT NULL,
    name VARCHAR(65535) NOT NULL
);

CREATE TABLE IF NOT EXISTS service_endpoint (
    id BIGSERIAL PRIMARY KEY,
    service_id BIGINT NOT NULL,
    code VARCHAR(65535) NOT NULL,
    name VARCHAR(65535) NOT NULL,
    UNIQUE (service_id, code),
    CONSTRAINT fk_service FOREIGN KEY (service_id) REFERENCES service (id) ON DELETE CASCADE ON UPDATE RESTRICT
);

CREATE TABLE IF NOT EXISTS service_endpoint_dependency (
    id BIGSERIAL PRIMARY KEY,
    service_endpoint_id BIGINT NOT NULL,
    dependecy_service_endpoint_id BIGINT NOT NULL,
    UNIQUE (service_endpoint_id, dependecy_service_endpoint_id),
    CONSTRAINT fk_service_endpoint FOREIGN KEY (service_endpoint_id) REFERENCES service_endpoint (id) ON DELETE CASCADE ON UPDATE RESTRICT
);
