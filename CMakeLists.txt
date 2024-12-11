cmake_minimum_required(VERSION 3.16)

project(jbpf-protobuf)
include(ExternalProject)

add_definitions(-DPB_FIELD_32BIT)

execute_process(
    COMMAND bash ./init_submodules.sh
    WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
    RESULT_VARIABLE result
    OUTPUT_VARIABLE output
    ERROR_VARIABLE error
)

set(JBPF_EXPERIMENTAL_FEATURES ON CACHE BOOL "Enable jbpf experimental features" FORCE)
add_subdirectory(jbpf)

if (NOT PROJECT_IS_TOP_LEVEL)
  ######## Export header files for other projects ############
  set(JBPF_IO_HEADER_FILES ${JBPF_IO_HEADER_FILES} PARENT_SCOPE)
  set(JBPF_MEM_MGMT_HEADER_FILES ${JBPF_MEM_MGMT_HEADER_FILES} PARENT_SCOPE)
  set(JBPF_LIB_HEADER_FILES ${JBPF_LIB_HEADER_FILES} PARENT_SCOPE)
  set(JBPF_CK_HEADER_FILES ${JBPF_CK_INCLUDES} PARENT_SCOPE)
  set(JBPF_COMMON_HEADERS ${JBPF_COMMON_HEADERS} PARENT_SCOPE)
  set(JBPF_LCM_IPC_HEADER_FILES ${JBPF_LCM_IPC_HEADER_FILES} PARENT_SCOPE)
endif()
