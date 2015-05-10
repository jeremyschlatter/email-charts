FROM golang:1.4

RUN apt-get update && \
    apt-get install -y --no-install-recommends r-base

RUN apt-get install -y --no-install-recommends build-essential && \
    echo 'options(repos=structure(c(CRAN="http://cran.cnr.berkeley.edu/")))' > $HOME/.Rprofile && \
    echo 'install.packages("ggplot2")' | R --no-save && \
    echo 'install.packages("chron")' | R --no-save && \
    echo 'install.packages("scales")' | R --no-save && \
    apt-get remove -y build-essential

COPY . /go/src/github.com/jeremyschlatter/email-charts
WORKDIR /go/src/github.com/jeremyschlatter/email-charts
RUN go install .

ENV EMAIL_CHARTS_SMTP_PASSWORD clhkmgfaceivvemy

ENTRYPOINT ["/go/bin/email-charts", "-stderr"]

EXPOSE 8080
