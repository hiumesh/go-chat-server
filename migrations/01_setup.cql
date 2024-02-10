
create table if not exists messages (
  id uuid PRIMARY KEY,
  channel_id uuid,
  user_id uuid,
  body text,
  created_at timestamp,
  updated_at timestamp
);

create table if not exists users (
  id uuid PRIMARY KEY,
  email text
);

create table if not exists channel (
  id uuid PRIMARY KEY,
  avatar text,
  name text
);

create table if not exists channel_users (
  channel_id uuid,
  user_id uuid,
  PRIMARY KEY (channel_id, user_id)
);
