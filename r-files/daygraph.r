library(ggplot2)
library(chron)
library(scales)
argv <- commandArgs(trailingOnly = TRUE)
mydata = read.csv(argv[1])
png(argv[2], height=300, width=470)


dayFinder <- function(x) weekdays(as.Date("1970/1/1") + x)

the_data <- data.frame(days=dayFinder(mydata$day))
ranking<-factor(the_data$days, c("Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"))


z <- ggplot(
		the_data, 
		aes(x=ranking)) +
	geom_histogram(
		color="#A9203E") +	
	ggtitle(
		"Email By Day") +
	theme_bw() +
	theme(
		legend.position = "none", 
		axis.title.x = element_text(vjust=-.1),
		plot.title = element_text(vjust=1, face="bold")) +
	xlab("Day of Week") +
	ylab("Quantity") 



print(z)
dev.off()



