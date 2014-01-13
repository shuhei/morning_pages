var utils = {
  lineBreak: function (str) {
    return str.replace(/\r?\n/g, '<br />');
  },
  pad: function (num) {
    var str = num.toString();
    if (str.length === 1) {
      str = '0' + str;
    }
    return str;
  },
  dateString: function (y, m, d) {
    return [y, utils.pad(m + 1), utils.pad(d)].join('-');
  },
  dateToString: function (date) {
    return [date.getFullYear(), date.getMonth() + 1, date.getDate()].map(utils.pad).join('-');
  },
  parseDate: function (str) {
    var cs = str.split('-').map(function (c) { return parseInt(c, 10); });
    return new Date(cs[0], cs[1] - 1, cs[2]);
  },
  beginningOfMonth: function (date) {
    return new Date(date.getFullYear(), date.getMonth(), 1);
  },
  endOfMonth: function (date) {
    return new Date(date.getFullYear(), date.getMonth() + 1, 0, 23, 59, 59, 999);
  },
  prevDate: function (date) {
    return new Date(date.getFullYear(), date.getMonth(), date.getDate() - 1);
  },
  nextDate: function (date) {
    return new Date(date.getFullYear(), date.getMonth(), date.getDate() + 1);
  },
  extractDay: function (dateString) {
    return dateString.split('-')[2].replace(/^0/, '');
  }
};

module.exports = utils;
