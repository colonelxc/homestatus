
#include <M5EPD.h>
#include <WiFi.h>
#include <HTTPClient.h>
#include "deserialize.h"
#include "config.h"


void setup(){
    M5.begin(/* touchEnable= */ false, /* SDEnable= */ false, /* SerialEnable= */ true, /* BatteryADCEnable= */ true, /* I2CEnable= */ false);
    M5.EPD.SetRotation(90);
    M5.RTC.begin();
    M5.SHT30.Begin();
    WiFi.mode(WIFI_STA);
    WiFi.begin(WIFI_SSID, WIFI_PASSWORD);
    while (!WiFi.isConnected()) {
      delay(1000);
    }
}

/* Render weather data. */
void handleWeather(tsv::Section& section, M5EPD_Canvas& canvas) {
  //canvas.drawString(String(section.GetSectionName().c_str()), 50, 20);
  
  std::vector<std::string> columnNames = section.ColumnNames();
  int namePos = -1;
  int temperaturePos = -1;
  int shortForecastPos = -1;
  int windSpeedPos = -1;
  for (int i = 0; i < columnNames.size(); i++) {
    std::string curr = columnNames.at(i);
    if (curr == "Name") {
      namePos = i;
    } else if (curr == "Temperature") {
      temperaturePos = i;
    } else if (curr == "ShortForecast") {
      shortForecastPos = i;
    } else if (curr == "WindSpeed") {
      windSpeedPos = i;
    }
  }
  if (namePos < 0 || temperaturePos < 0 || shortForecastPos < 0 || windSpeedPos < 0) {
    canvas.drawString("weather forecast missing columns", 0, 0);
    return;
  }

  int vpos=40;
  for (int i = 0; i < 4 && section.HasNextRow(); i++) {
    tsv::DataRow row = section.GetNextRow();

    canvas.setTextSize(4);
    canvas.drawString(String((row.GetColumn(temperaturePos) + " " + row.GetColumn(namePos)).c_str()), 15, vpos);
    canvas.setTextSize(3);
    canvas.drawString(row.GetColumn(shortForecastPos).c_str(), 30, vpos + 50);
    canvas.drawString("Wind speed: " + String(row.GetColumn(windSpeedPos).c_str()), 30, vpos + 80);
    vpos+=200;
  }
}

/* Render update time information. */
void handleTime(tsv::Section& section, M5EPD_Canvas& canvas) {
  int vpos = 900;
  canvas.setTextSize(2);
  std::vector<std::string> columnNames = section.ColumnNames();
  int lastUpdatedPos = -1;
  int currTimePos = -1;
  for (int i = 0; i < columnNames.size(); i++) {
    std::string curr = columnNames.at(i);
    if (curr == "lastUpdated") {
      lastUpdatedPos = i;
    } else if (curr == "currentTime") {
      currTimePos = i;
    }
  }
  if (lastUpdatedPos < 0 || currTimePos < 0) {
    canvas.drawString("UpdateTime missing columns", 50, 500);
    return;
  }
  tsv::DataRow row = section.GetNextRow();
  if (row.HasErr()) {
    
    canvas.drawString(row.Err().c_str(), 50, 500);
    return;
  }
  canvas.drawString("Data: " + String(row.GetColumn(lastUpdatedPos).c_str()), 15, vpos);
  canvas.drawString("Screen: " + String(row.GetColumn(currTimePos).c_str()), 15, vpos+20);
}

void loop(){
  M5.EPD.Clear(true); // Clear the screen.
  M5EPD_Canvas canvas(&M5.EPD);
  canvas.createCanvas(540, 960);
  canvas.setTextSize(3);

  HTTPClient http;
  http.begin(DATA_URL);
  int httpCode = http.GET();
  if (httpCode != 200) {
    char codeAsString[10];
    itoa(httpCode, codeAsString, 10);
    canvas.drawString("Http status code: " + String(codeAsString), 100, 50);
    http.end();
    delay(15 * 1000); // Wait a few seconds before running this loop again.
    return;
  }
  std::string data = std::string(http.getString().c_str());
  http.end();

  tsv::Deserializer deserializer = tsv::Deserializer::parse(data);

  while (deserializer.HasErr() == false && deserializer.HasNextSection()) {
    tsv::Section section = deserializer.GetNextSection();
    if (section.GetSectionName() == "WeatherForecast") {
      handleWeather(section, canvas);
    } else if (section.GetSectionName() == "UpdateTime") {
      handleTime(section, canvas);
    }
  }

  uint32_t mv = M5.getBatteryVoltage();
  char voltStr[10];
  itoa(mv, voltStr, 10);  

  canvas.setTextSize(2);
  canvas.drawString("Battery: " + String(voltStr) + "mV", 15, 880);
  
  
  canvas.pushCanvas(0, 0, UPDATE_MODE_A2);
  delay(1000); // Wait for the canvas to finish updating.

  M5.shutdown(3600);
  delay(5 * 60*1000); // If we are on usb power, it wont actually shutdown, so we just wait a few minutes before refreshing.
}
