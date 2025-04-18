-- +goose Up 
CREATE TABLE chirp(
   id UUID,
   created_at TIMESTAMP NOT NULL,
   updated_at TIMESTAMP NOT NULL,
   body TEXT NOT NULL,
   user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
   UNIQUE (user_id)
);

-- +goose Down
DROP TABLE chirp;

