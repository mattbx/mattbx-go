-- Tags for posts, same comma-separated shape as projects.tech. Added to hold
-- the mf2 category[] property from Micropub-created posts.
ALTER TABLE posts ADD COLUMN tags TEXT NOT NULL DEFAULT '';
