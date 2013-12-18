library(ggplot2)
library(chron)
library(scales)
argv <- commandArgs(trailingOnly = TRUE)
mydata = read.csv(argv[1])
png(argv[2], height=300, width=470)
	
	
# Takes time in seconds from midnight, converts to HH:MM:SS
timeHMS_formatter <- function(x) {					
	h <- floor(x/3600)
	m <- floor(x %% 60)
	s <- round(60*(x %% 1))                   		# Round to nearest second
	lab <- sprintf('%02d:%02d', h, m, s) 			# Format the strings as HH:MM:SS
	lab <- gsub('^00:', '', lab)              		# Remove leading 00: if present
	lab <- gsub('^0', '', lab)                		# Remove leading 0 if present
	}
	

date_chron = as.Date(as.chron(mydata$day))
the_data <- data.frame(date = date_chron, time=mydata$seconds)

z= 	ggplot(
		the_data, 
		aes(x=date, y= time)) +
	geom_point(
		alpha=5/8, 
		size=1.75, 
		color="#A9203E") + 
 	scale_x_date(
		breaks = date_breaks('year'),
		labels = date_format('%Y')) +
	coord_cartesian(
		xlim=c(min(the_data$date),max(the_data$date)), 
		ylim=c(0, 86400)) +
	scale_y_continuous(label=timeHMS_formatter, breaks=seq(0, 86400, 14400)) +
	ggtitle("Email Sending Times") +
	theme_bw() +														
	theme(
		legend.position = "none", 
		axis.title.x = element_text(vjust=-.1),
		plot.title = element_text(vjust=1, face="bold")) +
	xlab("Date") +
	ylab("Time of Day")	


print(z)
dev.off()
