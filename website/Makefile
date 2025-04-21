all: public

public: $(shell find src -type f)
	rm -rfv build
	kask build -in src -out build -domain /