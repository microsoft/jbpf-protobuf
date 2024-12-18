cmake_minimum_required(VERSION 3.16)

project(jbpf-protobuf)
include(ExternalProject)

add_definitions(-DPB_FIELD_32BIT)

################ Setup env ##################

set(OUTPUT_DIR ${CMAKE_SOURCE_DIR}/out)
set(ENV{JBPFP_OUT_DIR} ${OUTPUT_DIR})

# Execute the setup script first
option(INITIALIZE_SUBMODULES "Initialize submodules" ON)
if(INITIALIZE_SUBMODULES)
    message(STATUS "Initializing submodules")
    execute_process(
        COMMAND bash ./init_submodules.sh
        WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
        RESULT_VARIABLE result
        OUTPUT_VARIABLE output
        ERROR_VARIABLE error
    )
    if(NOT result EQUAL 0)
        message(FATAL_ERROR "Initializing submodules failed with exit code ${result}")
    endif()
else()
    message(STATUS "Skipping initializing submodules")
endif()

set(JBPF_EXPERIMENTAL_FEATURES ON CACHE BOOL "Enable jbpf experimental features" FORCE)
add_subdirectory(jbpf)
add_subdirectory(pkg)

if (NOT PROJECT_IS_TOP_LEVEL)
  ######## Export header files for other projects ############
  set(JBPF_IO_HEADER_FILES ${JBPF_IO_HEADER_FILES} PARENT_SCOPE)
  set(JBPF_MEM_MGMT_HEADER_FILES ${JBPF_MEM_MGMT_HEADER_FILES} PARENT_SCOPE)
  set(JBPF_LIB_HEADER_FILES ${JBPF_LIB_HEADER_FILES} PARENT_SCOPE)
  set(JBPF_CK_HEADER_FILES ${JBPF_CK_INCLUDES} PARENT_SCOPE)
  set(JBPF_COMMON_HEADERS ${JBPF_COMMON_HEADERS} PARENT_SCOPE)
  set(JBPF_LCM_IPC_HEADER_FILES ${JBPF_LCM_IPC_HEADER_FILES} PARENT_SCOPE)
endif()
