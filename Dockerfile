FROM golang

ADD *.go /go/src/github.com/andir/UthgardCommunityHeraldBackend/
ADD timeseries /go/src/github.com/andir/UthgardCommunityHeraldBackend/timeseries

ADD vendor /go/src/github.com/andir/UthgardCommunityHeraldBackend/vendor

RUN go install github.com/andir/UthgardCommunityHeraldBackend

VOLUME "./data"

WORKDIR /

EXPOSE 8081
CMD /go/bin/UthgardCommunityHeraldBackend
