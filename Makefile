open:
	- rm notes/conf.svg
	plantuml -tsvg notes/conf.puml
	firefox-esr notes/conf.svg

svg:
	- rm notes/conf.svg
	plantuml -tsvg notes/conf.puml

png:
	- rm notes/conf.png
	plantuml notes/conf.puml
	firefox-esr notes/conf.png

open:
	- rm notes/conf.svg
	plantuml -tsvg notes/conf.puml
	firefox-esr notes/conf.svg

public:
	scp notes/conf.svg 8i8_jyo@ssh.phx.nearlyfreespeech.net:/home/protected/public

private:
	scp notes/conf.svg 8i8_jyo@ssh.phx.nearlyfreespeech.net:/home/protected/private

