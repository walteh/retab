FROM ubuntu:20.04
RUN apt-get update && apt-get install -y \
	python3 \
	python3-pip \
	git
COPY . /app
WORKDIR    /app
RUN pip3 install -r requirements.txt

EXPOSE   8000

CMD  [  "python3",   "app.py"   ] 