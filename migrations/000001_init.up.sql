CREATE TABLE roles (
  id varchar(8) PRIMARY KEY,
  name varchar(255) NOT NULL
);

-- 
-- Seed data to table: roles
--
INSERT INTO roles (id, name) VALUES ('001ADMID', 'Admin'),('002USER', 'User');


CREATE TABLE permissions (
  id varchar(8) PRIMARY KEY,
  action varchar(191) NOT NULL,
  resource varchar(191) NOT NULL
);

--
-- Seed data to table: permissions
--
INSERT INTO permissions (id, action, resource) VALUES ('PERM0001', 'create', 'enrollments');
INSERT INTO permissions (id, action, resource) VALUES ('PERM0002', 'read', 'enrollments');

CREATE TABLE role_policies (
  role_id varchar(8) NOT NULL,
  permission_id varchar(8) NOT NULL,
  PRIMARY KEY (role_id, permission_id)
);

--
-- Seed data to table: role_policies
--
INSERT INTO role_policies (role_id, permission_id) VALUES ('001ADMID', 'PERM0001');
INSERT INTO role_policies (role_id, permission_id) VALUES ('001ADMID', 'PERM0002');
INSERT INTO role_policies (role_id, permission_id) VALUES ('002USER', 'PERM0002');

CREATE TABLE users (
  id varchar(8) PRIMARY KEY,
  name varchar(255) NOT NULL,
  email varchar(255) NOT NULL,
  role_id varchar(8) NOT NULL,
  FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE SET NULL
);

-- 
-- Seed data to table: users
-- 
INSERT INTO users (id, name, email, role_id) VALUES ('LITD0001', 'John Doe', 'john.litd@laoitdev.com','001ADMID');
INSERT INTO users (id, name, email, role_id) VALUES ('LITD0002', 'Snow Highway', '--', '002USER');

