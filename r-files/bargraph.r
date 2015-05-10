library(ggplot2)
library(chron)
library(scales)	
argv <- commandArgs(trailingOnly = TRUE)
mydata = read.csv(argv[1])
png(argv[2], height=300, width=470)


date_chron = as.Date(as.chron(mydata$day))
the_data <- data.frame(date = date_chron, seconds = mydata$seconds)
	
z <- ggplot(
		the_data, 
		aes(x=date)) +
	geom_histogram(
		color="#A9203E",
		binwidth=60) +	
	scale_x_date(
		breaks = date_breaks('year'),
		labels = date_format('%Y')) +
	scale_y_continuous(
		breaks=seq(0, 1000, 50)) + #define relative to max
	ggtitle(
		"Email Over Time") +
	theme_bw() +
	theme(
		legend.position = "none", 
		axis.title.x = element_text(vjust=-.1),
		plot.title = element_text(vjust=1, face="bold"),
		axis.text.y = element_blank(),
		axis.ticks.y = element_blank()) +
	xlab("Date") +
	ylab("Quantity") +
	coord_cartesian(
		xlim=c(min(the_data$date),max(the_data$date)))

print(z)
dev.off()


# Later - this should definitely be made into a line graph. It's completely and utterly misinformative as it is, because the binwidth *does* do what I thought it did. The graph currently shows that most email 









