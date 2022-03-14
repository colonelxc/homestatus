#include<string>
#include<vector>
#include<utility>
#include<memory>

namespace tsv {

  class Section;
  class DataRow;


  namespace {
    std::string extractString(std::string d, std::pair<size_t, size_t> p) {
      return d.substr(p.first, p.second);
    }

    std::vector<std::string> splitToVector(const std::string& data, std::string splitter) {
      std::vector<std::string> vec;
      size_t pos = 0;
      while(pos < data.length()) {
        size_t nextPos = data.find(splitter, pos);
        if (nextPos == std::string::npos) {
          vec.push_back(data.substr(pos, std::string::npos));
          break;
        } else {
          vec.push_back(data.substr(pos, nextPos-pos));
        }
        pos = nextPos + splitter.length();        
      }      
      return vec;
    }
  }


class Deserializer {
  public:
  static Deserializer parse(const std::string& data);

  bool HasNextSection();
  Section GetNextSection();

  bool HasErr();
  std::string Err();

  private:
  Deserializer(std::vector<std::string> sections, std::shared_ptr<std::string> err): 
    sections{sections}
    , err{err}
    , currSection{-1}
    {};

  std::vector<std::string> sections;
  std::shared_ptr<std::string> err;
  int currSection;
};

class Section {
  public:

  static Section parse(const std::string& data, const std::shared_ptr<std::string> err);
  std::string GetSectionName();
  std::vector<std::string> ColumnNames();
  bool HasNextRow();
  DataRow GetNextRow();
  bool HasErr();
  std::string Err();


  private:
  Section(const std::string sectionName, const std::vector<std::string> columns, const std::vector<std::string> rows, const std::shared_ptr<std::string> err)
    : sectionName{sectionName}
    , columns{columns}
    , rows{rows}
    , err{err}
    , currRow{-1}
    {};

  const std::string sectionName;
  const std::vector<std::string> columns;
  const std::vector<std::string> rows;
  const std::shared_ptr<std::string> err;
  int currRow;
};

class DataRow {
  public:  
    DataRow(const std::vector<std::string> columns, std::shared_ptr<std::string> err)
      : columns{columns}
      , err{err}
      {};

    /* Parse a string into a row of data. */
    static DataRow parse(const std::string& data, std::shared_ptr<std::string> err, int expectedColumns);
    /* Returns the column at index i. Sets Err if an invalid index is requested. */
    std::string GetColumn(int i);
    bool HasErr();
    std::string Err();

  private:
    const std::vector<std::string> columns;
    const std::shared_ptr<std::string> err;
};


}
