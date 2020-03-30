CREATE EXTENSION IF NOT EXISTS citext;

DROP TABLE IF EXISTS Users CASCADE;
DROP TABLE IF EXISTS Forum CASCADE;
DROP TABLE IF EXISTS Thread CASCADE;
DROP TABLE IF EXISTS Post CASCADE;
DROP TABLE IF EXISTS Vote CASCADE;
DROP TABLE IF EXISTS ForumUser CASCADE;

---------------------------------------------------------------------------

CREATE TABLE Users (
    nickname CITEXT UNIQUE NOT NULL PRIMARY KEY,
    fullname VARCHAR(255) NOT NULL,
    about TEXT,
    email CITEXT UNIQUE NOT NULL
);

---------------------------------------------------------------------------

CREATE TABLE Forum (
    title VARCHAR(255) NOT NULL,
    forumUser CITEXT REFERENCES Users(nickname) NOT NULL,
    slug CITEXT UNIQUE NOT NULL PRIMARY KEY,
    posts BIGINT default 0,
    threads BIGINT default 0
);

CREATE INDEX IF NOT EXISTS idx_forum_user ON Forum(forumUser);

---------------------------------------------------------------------------

CREATE TABLE Thread (
                        id BIGSERIAL PRIMARY KEY,
                        title VARCHAR(255) NOT NULL,
                        author CITEXT REFERENCES Users(nickname) NOT NULL,
                        forum CITEXT REFERENCES Forum(slug) NOT NULL,
                        message TEXT NOT NULL,
                        votes BIGINT default 0,
                        slug CITEXT UNIQUE,
                        created TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_thread_author ON Thread(author);
CREATE INDEX IF NOT EXISTS idx_thread_forum ON Thread(forum);

CREATE OR REPLACE FUNCTION updatethreadcount() RETURNS TRIGGER AS
$body$
BEGIN
    UPDATE Forum
    SET threads = threads + 1
    WHERE slug = NEW.forum;
    RETURN NEW;
END;
$body$ LANGUAGE plpgsql;

CREATE TRIGGER update_thread_count_trigger
    AFTER INSERT
    ON Thread
    FOR EACH ROW
EXECUTE PROCEDURE updatethreadcount();

CREATE OR REPLACE FUNCTION insertforumuser() RETURNS TRIGGER AS
$body$
BEGIN
    INSERT INTO ForumUser(slug, nickname)
    VALUES (NEW.forum, NEW.author)
    ON CONFLICT DO NOTHING;
    RETURN NEW;
END;
$body$ LANGUAGE plpgsql;

CREATE TRIGGER insert_forum_user_trigger_thread
    AFTER INSERT
    ON Thread
    FOR EACH ROW
EXECUTE PROCEDURE insertforumuser();

---------------------------------------------------------------------------

CREATE TABLE Post (
    id BIGSERIAL PRIMARY KEY,
    parent BIGINT default 0,
    author CITEXT REFERENCES Users(nickname) NOT NULL,
    message TEXT NOT NULL,
    isEdited BOOLEAN default FALSE,
    forum CITEXT REFERENCES Forum(slug) NOT NULL,
    thread BIGINT REFERENCES Thread(id) NOT NULL,
    created TIMESTAMP WITH TIME ZONE DEFAULT now(),
    path     BIGINT[] NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_post_author ON Post(author);
CREATE INDEX IF NOT EXISTS idx_post_forum ON Post(forum);
CREATE INDEX IF NOT EXISTS idx_post_thread ON Post(thread);

CREATE TRIGGER insert_forum_user_trigger_post
    AFTER INSERT
    ON Post
    FOR EACH ROW
EXECUTE PROCEDURE insertforumuser();

CREATE OR REPLACE FUNCTION updatepostcount() RETURNS TRIGGER AS
$body$
BEGIN
    UPDATE Forum
    SET posts = posts + 1
    WHERE slug = NEW.forum;
    RETURN NEW;
END;
$body$ LANGUAGE plpgsql;

CREATE TRIGGER update_post_count_trigger
    AFTER INSERT
    ON Post
    FOR EACH ROW
EXECUTE PROCEDURE updatepostcount();

CREATE OR REPLACE FUNCTION createpath() RETURNS TRIGGER AS
$postmatpath$
BEGIN
    NEW.path = (SELECT path FROM Post WHERE id = NEW.parent) || NEW.id;
    RETURN NEW;
END;
$postmatpath$ LANGUAGE plpgsql;

CREATE TRIGGER create_path_trigger
    BEFORE INSERT
    ON Post
    FOR EACH ROW
EXECUTE PROCEDURE createpath();

CREATE OR REPLACE FUNCTION updateiseditedcolumn() RETURNS TRIGGER AS
$body$
BEGIN
    IF NEW.message != OLD.message THEN
        NEW.isEdited = TRUE;
    END IF;
    RETURN NEW;
END;
$body$ LANGUAGE plpgsql;

CREATE TRIGGER update_isedited_column_trigger
    BEFORE UPDATE
    ON Post
    FOR EACH ROW
EXECUTE PROCEDURE updateiseditedcolumn();

---------------------------------------------------------------------------

CREATE TABLE Vote (
    id BIGSERIAL PRIMARY KEY,
    threadID BIGINT REFERENCES Thread(id) NOT NULL,
    author CITEXT REFERENCES Users(nickname) NOT NULL,
    voice SMALLINT NOT NULL,
    CONSTRAINT unique_vote UNIQUE (threadID, author)
);

CREATE INDEX IF NOT EXISTS idx_vote_threadid ON Vote(threadID);
CREATE INDEX IF NOT EXISTS idx_vote_author ON Vote(author);

CREATE OR REPLACE FUNCTION updatethreadvotes() RETURNS TRIGGER AS
$voteupdatecount$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        UPDATE Thread
        SET votes = votes + NEW.voice
        WHERE id = NEW.threadId;
    ELSE
        UPDATE Thread
        SET votes = votes - OLD.voice + NEW.voice
        WHERE id = NEW.threadId;
    END IF;
    RETURN NEW;
END;
$voteupdatecount$ LANGUAGE plpgsql;

CREATE TRIGGER update_thread_votes_trigger
    AFTER UPDATE OR INSERT
    ON Vote
    FOR EACH ROW
EXECUTE PROCEDURE updatethreadvotes();

---------------------------------------------------------------------------

CREATE TABLE ForumUser
(
    slug     CITEXT,
    nickname CITEXT,
    CONSTRAINT unique_slug_nickname UNIQUE (slug, nickname)
);

CREATE INDEX IF NOT EXISTS idx_forumUser_slug on ForumUser (slug);
CREATE INDEX IF NOT EXISTS idx_forumUser_nickname on ForumUser (nickname);