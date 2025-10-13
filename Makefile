## Makefile to format all Go files in all modules using 'go fmt'.

.PHONY: fmt
fmt:
	@echo "\033[1;32mFormatting Go files in all modules...\033[0m"
	@grep -E '^\s*\.\/' go.work | sed 's|^\s*./||' | while read dir; do \
		echo "\033[1;33mFormatting \033[1;34m[$$dir]\033[1;33m ...\033[0m"; \
		cd $$dir && go fmt ./... && cd ..; \
	done
	@echo "\033[1;32mFormatting complete!\033[0m"