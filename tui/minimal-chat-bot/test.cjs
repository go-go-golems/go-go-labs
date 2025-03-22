"use strict";

var _react = _interopRequireDefault(require("react"));
var _ink = require("ink");
function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { "default": obj }; }
var BasicExample = function BasicExample() {
  return /*#__PURE__*/_react["default"].createElement(_ink.Box, {
    flexDirection: "column",
    padding: 1
  }, /*#__PURE__*/_react["default"].createElement(_ink.Box, {
    marginBottom: 1
  }, /*#__PURE__*/_react["default"].createElement(_ink.Text, {
    bold: true,
    color: "#9D8CFF"
  }, "Hello from Ink!")), /*#__PURE__*/_react["default"].createElement(_ink.Box, null, /*#__PURE__*/_react["default"].createElement(_ink.Text, null, "This is a basic example of using Ink with React.")), /*#__PURE__*/_react["default"].createElement(_ink.Box, {
    marginTop: 1
  }, /*#__PURE__*/_react["default"].createElement(_ink.Text, {
    color: "#FF6B6B"
  }, "Press Ctrl+C to exit")));
};
(0, _ink.render)( /*#__PURE__*/_react["default"].createElement(BasicExample, null));
