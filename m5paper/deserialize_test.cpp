#include "test.h"
#include "deserialize.h"
#include<string>

TEST(test deserializer)(Testing& t) {
  const std::string data = "name\na\tb\tc\nx\ty\tz\n\nname2\nx\n\1\n2\n3\n\n";
  tsv::Deserializer deserializer = tsv::Deserializer::parse(data);

  if (deserializer.HasErr() == true) {
    t.failf("Err() = %s, want \"\"", deserializer.Err().c_str());
  }
  int totalSections = 0;
  int totalRows = 0;
  while(deserializer.HasNextSection()) {
    auto section = deserializer.GetNextSection();
    totalSections++;
    if (section.HasErr() == true) {
      t.failf("Err() = %s, want \"\"", section.Err().c_str());
    }
    while(section.HasNextRow()) {
      auto row = section.GetNextRow();
      totalRows++;
      if (row.HasErr()) {
        t.failf("Err() = %s, want \"\"", row.Err().c_str());
      }
      for (int i = 0; i < section.ColumnNames().size(); i++) {
        auto d = row.GetColumn(i);
        if (row.HasErr()) {
          t.failf("Err() = %s, want \"\"", row.Err().c_str());
        }
      }
    }
  }
  if (totalSections != 2) {
    t.failf("Got %d sections, wanted 2", totalSections);
  }
  if (totalRows != 4) {
    t.failf("Got %d rows, wanted 4", totalRows);
  }
}

TEST(test Section parse)(Testing& t) {
  const std::string data = "name\na\tb\tc\nx\ty\tz";
  std::shared_ptr<std::string> err = std::make_shared<std::string>("");
  tsv::Section section = tsv::Section::parse(data, err);

  if (section.HasErr() == true) {
    t.failf("Err() = %s, want \"\"", section.Err().c_str());
  }

  if (section.ColumnNames().size() != 3) {
    t.failf("ColumnNames().size() = %d, want 3", section.ColumnNames().size());
  }
  if (section.HasNextRow() != true) {
    t.fail("HasNextRow() = false, wanted true");
    return;
  }
  auto row = section.GetNextRow();
  if (row.GetColumn(2) != "z") {
    t.failf("GetColumn(2) = %s, wanted \"z\"", row.GetColumn(2));
  }
  if (row.GetColumn(1) != "y") {
    t.failf("GetColumn(1) = %s, wanted \"y\"", row.GetColumn(1));
  }
  if (row.GetColumn(0) != "x") {
    t.failf("GetColumn(0) = %s, wanted \"x\"", row.GetColumn(0));
  }
  if (section.HasNextRow() == true) {
    t.fail("HasNextRow() (2nd time) = true, wanted false");
  }
  if (section.HasErr() == true) {
    t.failf("HasErr() == true, wanted false");
  }
}

TEST(test section multiple rows)(Testing& t) {
  const std::string data = "name\na\nr\nr\nr\nr";
  std::shared_ptr<std::string> err = std::make_shared<std::string>("");
  tsv::Section section = tsv::Section::parse(data, err);
  int rowCount = 0;
  
  if (section.HasErr() == true) {
    t.failf("Err() = %s, want \"\"", section.Err().c_str());
  }
  while(section.HasNextRow()) {
    if (section.HasErr() == true) {
      t.failf("Err() = %s, want \"\"", section.Err().c_str());
    }
    auto row = section.GetNextRow();
    if (!row.HasErr() && row.GetColumn(0) != "r") {
      t.failf("GetColumn(0) = %s, wanted \"r\"", row.GetColumn(0).c_str());
    }
    rowCount++;

  }
  if (rowCount != 4) {
    t.failf("rowCount = %d, wanted 4", rowCount);
  }
}

TEST(test DataReader parse)(Testing& t) {
  const std::string data = "a\tb\tc";
  std::shared_ptr<std::string> err = std::make_shared<std::string>("");
  tsv::DataRow row = tsv::DataRow::parse(data, err, 3);


  if (row.HasErr()) {
	  t.failf("Err() = %s, want \"\"", row.Err().c_str());
  }
  std::string s = row.GetColumn(0);
  if (s != "a") {
    t.failf("row.GetColumn(0)=%s, want=a)", s.c_str());
  }
  s = row.GetColumn(1);
  if (s != "b") {
    t.failf("row.GetColumn(1)=%s, want=b)", s.c_str());
  }
  s = row.GetColumn(2);
  if (s != "c") {
    t.failf("row.GetColumn(2)=%s, want=c)", s.c_str());
  }
}

TEST(test DataRow too few cols)(Testing& t) {
  std::string data = "a\tb\tc";
  std::shared_ptr<std::string> err = std::make_shared<std::string>("");
  tsv::DataRow row = tsv::DataRow::parse(data, err, 10);


  if (!row.HasErr()) {
	  t.failf("HasErr() = %d, want 1", row.HasErr());
  }
}

TEST(test DataRow too many cols)(Testing& t) {
  std::string data = "a\tb\tc\td";
  std::shared_ptr<std::string> err = std::make_shared<std::string>("");
  tsv::DataRow row = tsv::DataRow::parse(data, err, 3);


  if (!row.HasErr()) {
	  t.failf("HasErr() = %d, want 1", row.HasErr());
  }
}

TEST(test DataRow GetColumn beyond col count)(Testing& t) {
  std::string data = "a\tb\tc";
  std::shared_ptr<std::string> err = std::make_shared<std::string>("");
  tsv::DataRow row = tsv::DataRow::parse(data, err, 3);

  // No error yet
  if (row.HasErr()) {
	  t.failf("HasErr() = %s, want \"\"", row.Err().c_str());
  }

  row.GetColumn(100);

  // Now error is expected
  if (!row.HasErr()) {
	  t.failf("HasErr() = %d, want 1", row.HasErr());
  }
}

TEST(test DataRow GetColumn negative)(Testing& t) {
  std::string data = "a\tb\tc";
  std::shared_ptr<std::string> err = std::make_shared<std::string>("");
  tsv::DataRow row = tsv::DataRow::parse(data, err, 3);

  // No error yet
  if (row.HasErr()) {
	  t.failf("HasErr() = %s, want \"\"", row.Err().c_str());
  }

  row.GetColumn(-1);

  // Now error is expected
  if (!row.HasErr()) {
	  t.failf("HasErr() = %d, want 1", row.HasErr());
  }
}