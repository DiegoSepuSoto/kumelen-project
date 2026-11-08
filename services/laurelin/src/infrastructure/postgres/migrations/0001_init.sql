CREATE TABLE institutions (
    id         uuid PRIMARY KEY,
    code       text NOT NULL UNIQUE,
    name       text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE programs (
    id             uuid PRIMARY KEY,
    institution_id uuid NOT NULL REFERENCES institutions (id),
    name           text NOT NULL,
    level          text NOT NULL CHECK (level IN ('undergraduate', 'master', 'doctorate', 'other_postgraduate')),
    created_at     timestamptz NOT NULL,
    updated_at     timestamptz NOT NULL,
    UNIQUE (institution_id, name, level)
);

-- Registro mínimo a propósito: sin RUT, teléfono ni dirección.
CREATE TABLE students (
    id             uuid PRIMARY KEY,
    institution_id uuid NOT NULL REFERENCES institutions (id),
    program_id     uuid NOT NULL REFERENCES programs (id),
    email          text NOT NULL UNIQUE CHECK (email = lower(email)),
    full_name      text NOT NULL,
    entry_year     smallint NOT NULL,
    status         text NOT NULL CHECK (status IN ('active', 'on_leave', 'graduated', 'withdrawn')),
    created_at     timestamptz NOT NULL,
    updated_at     timestamptz NOT NULL
);
CREATE INDEX students_institution_program_idx ON students (institution_id, program_id);

CREATE TABLE admins (
    id             uuid PRIMARY KEY,
    email          text NOT NULL UNIQUE CHECK (email = lower(email)),
    role           text NOT NULL CHECK (role IN ('super_admin', 'institution_admin')),
    institution_id uuid REFERENCES institutions (id),
    created_at     timestamptz NOT NULL,
    CHECK ((role = 'super_admin' AND institution_id IS NULL) OR (role = 'institution_admin' AND institution_id IS NOT NULL))
);

-- Solo identificadores: ni correos ni nombres. Sin claves foráneas para que el rastro sobreviva a los borrados.
CREATE TABLE audit_log (
    id             uuid PRIMARY KEY,
    actor_admin_id uuid NOT NULL,
    action         text NOT NULL,
    entity_type    text NOT NULL,
    entity_id      uuid NOT NULL,
    institution_id uuid,
    at             timestamptz NOT NULL
);
CREATE INDEX audit_log_at_idx ON audit_log (at);
