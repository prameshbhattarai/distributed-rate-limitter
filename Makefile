PROJECTS := $(wildcard */)

.PHONY: all build run clean

start: build run

# Build target
build:
	@for project in $(PROJECTS); do \
		if [ -f $$project/main.go ] || ls $$project/*.go > /dev/null 2>&1; then \
			echo "Building project in directory: $$project"; \
			cd $$project && go build -o ../$${project%/}_binary; \
			cd ..; \
		else \
			echo "No Go project found in directory: $$project"; \
		fi \
	done

# Run target
run:
	@echo "" > pids.txt
	@for project in $(PROJECTS); do \
		if [ -f $${project%/}_binary ]; then \
			echo "Running project: $${project%/}_binary"; \
			./$${project%/}_binary & \
			echo $$! >> pids.txt; \
		else \
			echo "No binary found for project: $$project"; \
		fi \
	done

# Stop target
stop:
	@echo "Stopping all running projects"
	@if [ -f pids.txt ]; then \
		for pid in $$(cat pids.txt); do \
			echo "Killing process $$pid"; \
			kill $$pid || true; \
		done; \
		rm -f pids.txt; \
	else \
		echo "No running projects found"; \
	fi


# Clean target to remove binaries
clean:
	@for project in $(PROJECTS); do \
		rm -f $${project%/}_binary; \
	done