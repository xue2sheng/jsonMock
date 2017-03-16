cmake_minimum_required(VERSION 3.0)

set(LOCAL_CMAKE_PROJECT_NAME JsonMock) 
project(${LOCAL_CMAKE_PROJECT_NAME} NONE) 

### pseudo install ###
set(BINARY_INSTALL_DIR ${CMAKE_CURRENT_BINARY_DIR})

### Code ###
add_subdirectory("src")

### Testing ###
if(${LOCAL_CMAKE_PROJECT_NAME}_TEST)
  add_subdirectory("test")
endif()
