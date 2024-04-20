#There are a bunch of SQL features that don't matter for SQLC but break their parser
sed -i -e 's/ON UPDATE NOW[(][)]//g' schema.sql
sed -i -e 's/CREATE DATABASE.*//g' schema.sql
sed -i -e 's/CREATE INDEX.*//g' schema.sql
sed -i -e 's/CREATE USER.*//g' schema.sql
sed -i -e 's/GRANT.*//g' schema.sql
sed -i -e 's/DESC//g' schema.sql
sed -i -e 's/USING HASH.*//g' schema.sql
rm -f query/*.sql.go
docker run --rm -v $(pwd):/src -w /src sqlc/sqlc:1.19.1 generate
rm *.sql-e
echo done