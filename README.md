Go Openweathermap
=================

This program fetches the current weather from [Open Weather Map](http://openweathermap.org/) and format it.

Formatting is done with a [Go template](http://golang.org/pkg/text/template/) using the data structure [main.CurrentWeather](https://github.com/vincent-petithory/go-openweathermap/blob/master/main.go#L34-L81).

### Usage




Here's an example which will print something like:

	Paris 25°C Scattered clouds

Code:

	echo '{{.Name}} {{temp .Main.Temp.ToC}}°C {{(index .Weather 0).Description}}' | go-openweathermap --once 2988507

A bit more complex with a weather unicode char:

	Paris 25°C Scattered clouds ☁

Code:

	echo '{{.Name}} {{temp .Main.Temp.ToC}}°C {{(index .Weather 0).Description}}|{{(index .Weather 0).Id}}' | go-openweathermap --fetch-delay 10m 2988507 \
	| while read line; do
		text=$(echo "$line" | cut -f1 -d'|')
		condition_code=$(echo "$line" | cut -f2 -d'|')
		icon='?'
		case $condition_code in
			200|201|202|210|211|212|221|230|231|232)
			# thunderstorm
			icon=☂
			;;
			300|301|302|310|311|312|313|314|321)
			# drizzle
			icon=☂
			;;
			500|501|502|503|504|511|520|521|522|531)
			# rain
			icon=☂
			;;
			600|601|602|611|612|615|616|620|621|622)
			icon=☃
			;;
			801|802|803|804)
			icon=☁
			;;
			800)
			icon=☀
			;;
			esac
		echo $text $icon
	done
