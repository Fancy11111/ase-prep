var express = require('express');
var router = express.Router();

/* GET home page. */
const base = '/:scenario/assignment/:mnr'
router.get(base + '/token', function (req, res, next) {
  console.log(req.params)
  res.send("token123")
});

router.get(base + '/stage/:stage/testcase/:testcase', function (req, res, next) {
  console.log(req.params)
  console.log(req.query["token"])
  res.json({
    obstacle: {
      line: -3,
      pointA: {x: 3, y: -3},
      pointB: {x: -3, y: -3}
    },
    targets: [
      {x: 10, y: -5},
      {x: -6, y: -9},
      {x: 0, y: -1}
    ]
  })
});

router.post(base + '/stage/:stage/testcase/:testcase', function (req, res) {
  console.log(req.params)
  console.log(req.query["token"])
  console.log(req.body)
  res.json({
    message: "Accepted",
    linkToNextTask: `http://localhost:3030/${req.params.scenario}/assignment/${req.params.mnr}/stage/${req.params.stage}/testcase/${+req.params.testcase + 1}?token=${req.query.token}`
  })
});

module.exports = router;
