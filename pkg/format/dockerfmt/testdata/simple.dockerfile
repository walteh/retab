FROM ubuntu:20.04
RUN apt-get update
RUN apt-get install -y python3
COPY . /app
WORKDIR /app
EXPOSE 8000
CMD ["python3", "app.py"] 