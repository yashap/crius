CREATE TABLE IF NOT EXISTS service (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(511) UNIQUE NOT NULL,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS service_endpoint (
    id BIGSERIAL PRIMARY KEY,
    service_id BIGINT NOT NULL,
    code VARCHAR(511) NOT NULL,
    name TEXT NOT NULL,
    UNIQUE (service_id, code),
    CONSTRAINT fk_service FOREIGN KEY (service_id) REFERENCES service (id) ON DELETE CASCADE ON UPDATE RESTRICT
);

CREATE TABLE IF NOT EXISTS service_endpoint_dependency (
    id BIGSERIAL PRIMARY KEY,
    service_endpoint_id BIGINT NOT NULL,
    dependency_service_endpoint_id BIGINT NOT NULL,
    UNIQUE (service_endpoint_id, dependency_service_endpoint_id),
    CONSTRAINT fk_service_endpoint FOREIGN KEY (service_endpoint_id) REFERENCES service_endpoint (id) ON DELETE CASCADE ON UPDATE RESTRICT,
    CONSTRAINT fk_dependency_service_endpoint FOREIGN KEY (dependency_service_endpoint_id) REFERENCES service_endpoint (id) ON DELETE RESTRICT ON UPDATE RESTRICT
);
