# Database Documentation
Simplebank database documentation

> NOTE: Every command should be run from root project directory

## Requirements
- [Node.js](https://nodejs.org/)
- [dbdocs](https://dbdocs.io/docs)
- [@dbms/cli](https://dbml.dbdiagram.io/cli/)
- (Optional) [Docker](https://www.docker.com/)
- (Optional) [Docker Compose](https://docs.docker.com/compose/)

## Login
1. Create Account https://dbdiagram.io
2. Login login using dbdocs [Detail Information here](https://dbdocs.io/docs)
    ```sh
    $ dbdocs login
    ```


### Generate documentation
1. Generate docs
    ```sh
    $ make db_doc
    ```

### Generate documentation when database updated
1. Make sure docker compose is running
    ```sh
    $ docker compose ps
    ```
    > Note: if there is no container running. Run `docker compose up -d`
2. Dump current database
    ```sh
    $ make db_dump
    ```
3. Convert .sql file into .dbml
    ```sh
    $ make db_dbml
    ```
    > Note: If there is any error, fix .sql file manually
4. Generate docs
    ```sh
    $ make db_doc
    ```
