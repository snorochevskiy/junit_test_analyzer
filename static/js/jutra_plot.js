
//+======================================================+
//|                                                      |
//|    (plotX0, plotY0)                                  |
//|      |AXIS_GAP                                       |
//|      |*                                  PLOT_X_GAP  |
//|      | *                                             |
//|      |  *                                            |
//|      |    *           ***                            |
//|      |      *        *    *               ****       |
//|      |        *     *      ***************           |
//|      |          ***                     AXIS_GAP     |
//|      +------------------------------------------     |
//|     0 TEXT_INDENT              PLOT_Y_GAP            |
//+======================================================+

var coords = [];

var AXIS_GAP = 20;
var PLOT_X_GAP = 20;
var PLOT_Y_GAP = 20;
var TEXT_INDENT = 10;

function Chart(canvasElement, testRuns) {
	var self = this;
		
	this.canvas = canvasElement;
	this.testRuns = testRuns;
	
	this.ctx = canv.getContext('2d');

	// top left coordinate of plot area 
	this.plotX0 = 20;
	this.plotY0 = 20;
	
	this.plotWidth = this.canvas.width - (this.plotX0 +  PLOT_X_GAP);
	this.plotHeight = this.canvas.height - (this.plotY0 + PLOT_Y_GAP);
	
	this.culculatePlots = function() {
		self.plotWidth = self.canvas.width - (self.plotX0 +  PLOT_X_GAP);
		self.plotHeight = self.canvas.height - (self.plotY0 + PLOT_Y_GAP);
	}
	
	this.culculateCoords = function() {
		// Find axis multipliers
		var maxX = testRuns[0].date;
		var minX = testRuns[0].date;
		var maxY = testRuns[0].fails;
		var minY = testRuns[0].fails;
		var plotWidth = self.plotWidth;
		var plotHeight = self.plotHeight;
		var plotX0 = self.plotX0;
		var plotY0 = self.plotY0;

		for (i=0; i < testRuns.length; i++) {
			if (maxX < testRuns[i].date) {
				maxX = testRuns[i].date;
			}
			if (minX > testRuns[i].date) {
				minX = testRuns[i].date;
			}
			if (maxY < testRuns[i].fails) {
				maxY = testRuns[i].fails;
			}
			if (minY > testRuns[i].fails) {
				minY = testRuns[i].fails;
			}
		}
		
		self.minX = minX;
		self.minY = minY;
		
		var xMult = (plotWidth - AXIS_GAP) / (maxX - minX);
		var yMult = (plotHeight - AXIS_GAP) / (maxY - minY);
	
		self.xMult = xMult;
		self.yMult = yMult;
	
		var coords = []; 
		for (i=0; i < testRuns.length; i++) {
			coords.push({
					x: self.xCoord(testRuns[i].date),
					y: self.yCoord(testRuns[i].fails),
					time: new Date(testRuns[i].date * 1000) // Date from Unix time
				});
		}

		self.coords = coords;	
	}
	
	this.xCoord = function(date) {
		return Math.ceil(self.plotX0 + (date - self.minX) * self.xMult)
	}

	this.yCoord = function(failsNum) {
		return Math.ceil(self.plotY0 + self.plotHeight - (failsNum - self.minY) * self.yMult)
	}
	
	this.repaint = function () {
		self.culculatePlots();
		self.drawPlot();
		
		self.culculateCoords();
		self.drawMonthesBlocks();
		self.drawErrorLines();
		self.drawChart();
	}
	
	
	this.drawPlot = function() {
		var ctx = self.ctx;
		var plotX0 = self.plotX0;
		var plotY0 = self.plotY0;
		var plotWidth = self.plotWidth;
		var plotHeight = self.plotHeight;
		
		ctx.fillStyle = "#E8E8E8";
		ctx.fillRect(plotX0, plotY0, plotWidth, plotHeight);

		// Draw axis
		ctx.fillStyle = "black";
		ctx.beginPath();
		ctx.moveTo(plotX0, plotY0 + plotHeight);
		ctx.lineTo(plotX0, plotY0);
		ctx.moveTo(plotX0, plotY0 + plotHeight);
		ctx.lineTo(plotX0 + plotWidth, plotY0 + plotHeight);
		ctx.stroke();

		ctx.font="10px Arial";
		ctx.fillText("0", plotX0 - 2 * TEXT_INDENT, plotX0 + plotHeight + TEXT_INDENT);
	}
	
	this.drawMonthesBlocks = function() {
		var coords = self.coords;
		var testRuns = self.testRuns;
		var ctx = self.ctx;
		var plotX0 = self.plotX0;
		var plotY0 = self.plotY0;
		var plotWidth = self.plotWidth;
		var plotHeight = self.plotHeight;
		
		// Don't draw day lines if it's too much days
		var numberOfDays = (coords[coords.length - 1].time.getFullYear() - coords[0].time.getFullYear()) * 365 + 
				(coords[coords.length - 1].time.getMonth() - coords[0].time.getMonth()) * 30 +
				(coords[coords.length - 1].time.getDate() - coords[0].time.getDate());		
		var displayDaysGrid = (plotWidth / numberOfDays) > 3;
		
		if (displayDaysGrid) {
			
			var prevDate = new Date(coords[0].time.getFullYear(), coords[0].time.getMonth(), 1);
			var lastDate = coords[coords.length - 1].time;
			
			var firstDayLine = true;
			
			ctx.strokeStyle = "#8C8C8C";

			var initX = self.xCoord(coords[0].time.getTime() / 1000);
			var dateString = coords[0].time.toDateString().split(' ');
			ctx.fillText(dateString[1] + " " + dateString[3], initX, plotY0 + plotHeight + 2*TEXT_INDENT);

			for (var i = 0;;i++) {
				var nextDate = new Date(prevDate);
				nextDate.setDate(nextDate.getDate() + 1);
				
				if ((nextDate > lastDate) == 1) {
					break;
				}
				
				var isMonthChange = nextDate.getMonth() != prevDate.getMonth();
				
				var x = self.xCoord(nextDate.getTime() / 1000);
				ctx.setLineDash(isMonthChange ? [] : [5, 5]);
				ctx.beginPath();
				ctx.moveTo(x, plotY0 + plotHeight);
				ctx.lineTo(x, plotY0);
				ctx.stroke();
				
				ctx.fillText("" + nextDate.getDate(), x, plotY0 + plotHeight + TEXT_INDENT);
				if (isMonthChange) {
					var dateString = nextDate.toDateString().split(' '); // Sun. 31 Jan 2016 21:00:00 GMT
					ctx.fillText(dateString[1] + " " + dateString[3], x, plotY0 + plotHeight + 2*TEXT_INDENT);
				}
				
				prevDate = nextDate;
			}
			ctx.setLineDash([]);
		}
		
	}
	
	this.drawErrorLines = function() {
		var ctx = self.ctx;
		var coords = self.coords;
		var plotX0 = self.plotX0;
		var plotY0 = self.plotY0;
		var plotWidth = self.plotWidth;
		var plotHeight = self.plotHeight;
		
		ctx.strokeStyle = "#FFA54F";
		ctx.lineWidth=1;
		for (i = 1; i < coords.length; i++) {
			if (coords[i].y == plotY0 + plotHeight) {
				continue;
			}
			ctx.beginPath();
			ctx.moveTo(plotX0, coords[i].y);
			ctx.lineTo(plotX0 + plotWidth, coords[i].y);
			ctx.stroke();
		}
	}
	
	this.drawChart = function() {
		var ctx = self.ctx;
		var coords = self.coords;
		var testRuns = self.testRuns;
		var plotX0 = self.plotX0;
		var plotY0 = self.plotY0;
		var plotWidth = self.plotWidth;
		var plotHeight = self.plotHeight;
		var xMult = self.xMult;
		var yMult = self.yMult;
		
		// Draw chart
		ctx.fillStyle = "black";
		ctx.strokeStyle = "#b30000";//"#43CD80";
		ctx.lineWidth=3;
		ctx.beginPath();
		ctx.moveTo(coords[0].x, coords[0].y);
		for (i=1; i < testRuns.length; i++) {
			ctx.lineTo(coords[i].x, coords[i].y);
		}
		ctx.stroke();

		// Draw axis marks
		ctx.lineWidth=1;

		for (i=0; i < testRuns.length; i++) {

			var yText = testRuns[i].displayY == undefined ? "" + testRuns[i].fails : testRuns[i].displayY;
			if (yText != "0") {
				// Number of failed test at Y axis
				ctx.fillText(yText, plotX0 - 2 * TEXT_INDENT, plotY0 + (plotHeight - testRuns[i].fails * yMult));
			}
		}

	}
}


function plotHover(event) {
	
	for (var i = 0; i < coords.length; i++) {
		if ( Math.abs(coords[i].x - (event.clientX - 20)) < 5 && Math.abs(coords[i].y - (event.clientY - 30)) < 5) {
			// Clicked on run vertex
		}
	}
}