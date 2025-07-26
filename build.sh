builtAt="$(date +'%F %T %z')"
gitCommit=$(git log --pretty=format:"%h" -1)
version=$(git describe --abbrev=0 --tags)

ldflags="\
-w -s \
-X 'github.com/krau/ManyACG/common.BuildTime=$builtAt' \
-X 'github.com/krau/ManyACG/common.Commit=$gitCommit' \
-X 'github.com/krau/ManyACG/common.Version=$version'\
"

go build -ldflags "$ldflags" -o manyacg