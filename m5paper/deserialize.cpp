#include "deserialize.h"

/** Top level Deserializer */

tsv::Deserializer tsv::Deserializer::parse(const std::string& data) {
  std::vector<std::string> sections = splitToVector(data, "\n\n");
  std::string err = "";
  if (sections.size() == 0 || data.size() == 0) {
    err = "No sections were found in the data";
  }
  return Deserializer(sections, std::make_shared<std::string>(err));
}

bool tsv::Deserializer::HasNextSection() {
  if (HasErr()) {
    return false;
  }
  if ((currSection + 1) >= sections.size()) {
    return false;
  }
  return true;
}

tsv::Section tsv::Deserializer::GetNextSection() {
  if (HasErr()) {
    return tsv::Section::parse("", err);
  }
  if (!HasNextSection()) {
    err->assign("Tried to get the next section, but there is none");
    return tsv::Section::parse("", err);
  }
  currSection++;
  return tsv::Section::parse(sections.at(currSection), err);
}

bool tsv::Deserializer::HasErr() {
  return !err->empty();
}

std::string tsv::Deserializer::Err() {
  return *err.get();
}

/** Section Parsing */
tsv::Section tsv::Section::parse(const std::string& data, std::shared_ptr<std::string> err) {
  std::vector<std::string> sectionData;
  if (!err->empty()) {
    // Return a dummy section, with the error.
    return Section("", std::vector<std::string>(), std::vector<std::string>(), err);
  }
  sectionData = splitToVector(data, "\n");
  if (sectionData.size() < 3) {
    err->assign("Not enough data to form a full section");
  }
  std::string sectionName = sectionData.at(0);
  if (sectionName.size() == 0) {
    err->assign("empty section name");
  }
  std::vector<std::string> columns = splitToVector(sectionData.at(1), "\t");
  if (columns.size() == 0) {
    err->assign("No columns in this section");
  }

  std::vector<std::string> rows;
  rows.reserve(sectionData.size() - 2);
  for (int i = 2; i < sectionData.size(); i++) {
    rows.push_back(sectionData.at(i));
  }
  return Section(sectionName, columns, rows, err);
}

std::string tsv::Section::GetSectionName() {
  if (HasErr()) {
    return "";
  }
  return sectionName;
}

bool tsv::Section::HasNextRow() {
  if (HasErr()) {
    return false;
  }
  if ((currRow + 1) >= rows.size()) {
    return false;
  }
  return true;
}

tsv::DataRow tsv::Section::GetNextRow() {
  if (!HasNextRow()) {
    return tsv::DataRow::parse("", err, 0);
  }
  currRow++;
  return tsv::DataRow::parse(rows.at(currRow), err, columns.size());
}

std::vector<std::string> tsv::Section::ColumnNames() {
  if (HasErr()) {
    return std::vector<std::string>();
  }
  return columns;
}

bool tsv::Section::HasErr() {
  return !err->empty();
}

std::string tsv::Section::Err() {
  return *err.get();
}

/** DataRow Parsing */
tsv::DataRow tsv::DataRow::parse(const std::string& data, std::shared_ptr<std::string> err, int expectedColumns) {
  std::vector<std::string> columns = splitToVector(data, "\t");
  if (!err->empty()) {
    return DataRow(std::vector<std::string>(), err);
  }
  if (columns.size() != expectedColumns) {
    err->assign("Incorrect number of columns for row");
  }
  return DataRow(columns, err);
}

std::string tsv::DataRow::GetColumn(int i) {
  if (HasErr()) {
    return "";
  }
  if (i < 0 || i >= columns.size()) {
    err->assign("Column request out of bounds");
    return "";
  }
  return columns.at(i);
}

bool tsv::DataRow::HasErr() {
  return !err->empty();
}

std::string tsv::DataRow::Err() {
  return *err.get();
}
