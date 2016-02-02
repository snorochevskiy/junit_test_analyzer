
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

function DrawChart(canv, data) {
	var AXIS_GAP = 20;
	var PLOT_X_GAP = 20;
	var PLOT_Y_GAP = 20;
	var TEXT_INDENT = 10;

	var plotX0 = 20;
	var plotY0 = 20;
	var plotWidth = canv.width - (plotX0 +  PLOT_X_GAP);
	var plotHeight = canv.height - (plotY0 + PLOT_Y_GAP);

	var ctx = canv.getContext('2d');

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
	ctx.fillText("0", plotX0, plotX0 + plotHeight + TEXT_INDENT);

	// Find axis multipliers
	var maxX = dots[0].x;
	var minX = dots[0].x;
	var maxY = dots[0].y;
	var minY = dots[0].y;
	for (i=0; i < dots.length; i++) {
		if (maxX < dots[i].x) {
			maxX = dots[i].x;
		}
		if (minX > dots[i].x) {
			minX = dots[i].x
		}
		if (maxY < dots[i].y) {
			maxY = dots[i].y;
		}
		if (minY > dots[i].y) {
			minY = dots[i].y
		}
	}
	var xMult = (plotWidth - AXIS_GAP) / (maxX - minX);
	var yMult = (plotHeight - AXIS_GAP) / (maxY - minY);

	var coords = [];
	for (i=0; i < dots.length; i++) {
		coords.push({
				x:plotX0 + (dots[i].x - minX) * xMult,
				y:plotY0 + plotHeight - (dots[i].y - minY) * yMult
			});
	}

	ctx.strokeStyle = "#FFA54F";
	for (i=1; i < coords.length; i++) {
		if (coords[i].y == plotY0 + plotHeight) {
			continue;
		}
		ctx.beginPath();
		ctx.moveTo(plotX0, coords[i].y);
		ctx.lineTo(plotX0 + plotWidth, coords[i].y);
		ctx.stroke();
	}

	// Draw chart
	ctx.strokeStyle = "#43CD80";
	ctx.lineWidth=3;
	ctx.beginPath();
	ctx.moveTo(coords[0].x, coords[0].y);
	for (i=1; i < dots.length; i++) {
		ctx.lineTo(coords[i].x, coords[i].y);
	}
	ctx.stroke();

	// Draw axis marks
	ctx.lineWidth=1;
	//var level = 1;
	for (i=0; i < dots.length; i++) {
		//level = level == 4 ? 1 : level + 1;
		var xText = dots[i].displayX == undefined ? "" + dots[i].x : dots[i].displayX;
		if (xText != "0") {
			//ctx.fillText(xText, plotX0 + ((dots[i].x - minX) * xMult), plotY0 + plotHeight + level * TEXT_INDENT);
			ctx.fillText(xText, coords[i].x, coords[i].y);
		}

		var yText = dots[i].displayY == undefined ? "" + dots[i].y : dots[i].displayY;
		if (yText != "0") {
			ctx.fillText(yText, plotX0 - TEXT_INDENT, plotY0 + (plotHeight - dots[i].y * yMult));
		}
	}
}