find_program(LOCAL_GO_COMPILER go HINTS /home /usr)
message(STATUS "Go compiler: ${LOCAL_GO_COMPILER}")

if(EXISTS ${LOCAL_GO_COMPILER})

	if(WIN32)
		set(BINARY_EXE ".exe")
	else()
		set(BINARY_EXE "")
	endif()

	add_custom_target(${LOCAL_CMAKE_PROJECT_NAME} ALL ${LOCAL_GO_COMPILER} build ${CMAKE_CURRENT_SOURCE_DIR}/JsonMock.go COMMAND ${CMAKE_COMMAND} -E copy_directory ${CMAKE_CURRENT_SOURCE_DIR}/../data ${CMAKE_CURRENT_BINARY_DIR}/data)

 ### Only if this the principal project ###
 if("${LOCAL_CMAKE_PROJECT_NAME}" STREQUAL "${CMAKE_PROJECT_NAME}")
	 add_custom_target(install${LOCAL_CMAKE_PROJECT_NAME} ${CMAKE_COMMAND} -E copy_if_different ${CMAKE_CURRENT_BINARY_DIR}/${LOCAL_CMAKE_PROJECT_NAME}${BINARY_EXE} ${BINARY_INSTALL_DIR} COMMAND ${CMAKE_COMMAND} -E copy_directory ${CMAKE_CURRENT_BINARY_DIR}/data ${BINARY_INSTALL_DIR}/data DEPENDS ${LOCAL_CMAKE_PROJECT_NAME})
 endif()

 ### Testing ###
 if(${LOCAL_CMAKE_PROJECT_NAME}_TEST)
	 add_custom_target(${LOCAL_CMAKE_PROJECT_NAME}.test ${LOCAL_GO_COMPILER} test -c ${CMAKE_CURRENT_SOURCE_DIR}/JsonMock.test.go COMMAND ${CMAKE_COMMAND} -E copy_directory ${CMAKE_CURRENT_SOURCE_DIR}/../data ${CMAKE_CURRENT_BINARY_DIR}/data DEPENDS ${LOCAL_CMAKE_PROJECT_NAME})

   ### Only if this the principal project ###
   if("${LOCAL_CMAKE_PROJECT_NAME}" STREQUAL "${CMAKE_PROJECT_NAME}")
	   add_custom_target(install${LOCAL_CMAKE_PROJECT_NAME}.test ${CMAKE_COMMAND} -E copy_if_different ${CMAKE_CURRENT_BINARY_DIR}/${LOCAL_CMAKE_PROJECT_NAME}.test${BINARY_EXE} ${BINARY_INSTALL_DIR} COMMAND ${CMAKE_COMMAND} -E copy_directory ${CMAKE_CURRENT_BINARY_DIR}/data ${BINARY_INSTALL_DIR}/data DEPENDS install${LOCAL_CMAKE_PROJECT_NAME} ${LOCAL_CMAKE_PROJECT_NAME}.test)
	   add_custom_target(test${LOCAL_CMAKE_PROJECT_NAME} ./${LOCAL_CMAKE_PROJECT_NAME}.test${BINARY_EXE} DEPENDS install${LOCAL_CMAKE_PROJECT_NAME}.test WORKING_DIRECTORY ${BINARY_INSTALL_DIR})
   endif()

 endif()

else(EXISTS ${LOCAL_GO_COMPILER})

  add_custom_target(${LOCAL_CMAKE_PROJECT_NAME} ALL echo "No golang compiler means no ${LOCAL_CMAKE_PROJECT_NAME}")

  ### Testing ###
  if(${LOCAL_CMAKE_PROJECT_NAME}_TEST)
      add_custom_target(${LOCAL_CMAKE_PROJECT_NAME}.test echo "No golang compiler means no ${LOCAL_CMAKE_PROJECT_NAME}.test" DEPENDS ${LOCAL_CMAKE_PROJECT_NAME})
  endif()

endif(EXISTS ${LOCAL_GO_COMPILER})
