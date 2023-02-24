CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE "products" (
    "id" uuid NOT NULL DEFAULT uuid_generate_v4(),
    "name" text NOT NULL,
    "description" text NOT NULL,
    "category" text NOT NULL,
    "price" FLOAT NOT NULL,
    "quantity" INT NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT (now()),
    "updated_at" timestamptz NOT NULL DEFAULT (now()),
    PRIMARY KEY (id)
);
