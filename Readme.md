to run PostgreSQL in the docker container type in the command line:

docker run --name some-postgres -e POSTGRES_PASSWORD=gobank -p5432:5432 -d postgres