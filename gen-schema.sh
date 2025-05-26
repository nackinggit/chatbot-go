dsn='root:123456@tcp(127.0.0.1:3306)/chatbot?timeout=5s&readTimeout=5s&writeTimeout=5s&parseTime=true&loc=Local&charset=utf8,utf8mb4'
echo $dsn
gentool -dsn $dsn -onlyModel -outPath ./internal/model/db -modelPkgName dbmodel