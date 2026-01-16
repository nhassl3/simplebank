CREATE TABLE "admins" (
                          "owner" varchar NOT NULL,
                          "level_right" NUMERIC NOT NULL DEFAULT 1
);

CREATE INDEX ON "admins" ("owner");

ALTER TABLE "admins" ADD FOREIGN KEY ("owner") REFERENCES "users" ("username");
