/* Do not edit this file manually.
   This file is autogenerated by windmc.  */

//
//  The values are 32 bit layed out as follows:
//
//   3 3 2 2 2 2 2 2 2 2 2 2 1 1 1 1 1 1 1 1 1 1
//   1 0 9 8 7 6 5 4 3 2 1 0 9 8 7 6 5 4 3 2 1 0 9 8 7 6 5 4 3 2 1 0
//  +---+-+-+-----------------------+-------------------------------+
//  |Sev|C|R|     Facility          |               Code            |
//  +---+-+-+-----------------------+-------------------------------+
//
//  where
//
//      C    - is the Customer code flag
//
//      R    - is a reserved bit
//
//      Code - is the facility's status code
//
//      Sev  - is the severity code
//
//           Success - 00
//           Informational - 01
//           Warning - 02
//           Error - 03
//
//      Facility - is the facility code
//
//           System - 00ff
//           Application - 0fff
//

// Header
// Messages
//
// MessageId: MSG_WARN_REGCONFIG_FAILED
//
#define MSG_WARN_REGCONFIG_FAILED  0x80000001

//
// MessageId: MSG_WARN_CONFIGUPGRADE_FAILED
//
#define MSG_WARN_CONFIGUPGRADE_FAILED  0x80000002

//
// MessageId: MSG_SERVICE_STARTED
//
#define MSG_SERVICE_STARTED  0x40000003

//
// MessageId: MSG_SERVICE_STOPPED
//
#define MSG_SERVICE_STOPPED  0x40000004

//
// MessageId: MSG_UNKNOWN_CONTROL_REQUEST
//
#define MSG_UNKNOWN_CONTROL_REQUEST  0x80000005

//
// MessageId: MSG_SERVICE_STOPPING
//
#define MSG_SERVICE_STOPPING  0x40000006

//
// MessageId: MSG_SERVICE_STARTING
//
#define MSG_SERVICE_STARTING  0x40000007

//
// MessageId: MSG_SERVICE_FAILED
//
#define MSG_SERVICE_FAILED  0xc0000008

//
// MessageId: MSG_UNEXPECTED_CONTROL_REQUEST
//
#define MSG_UNEXPECTED_CONTROL_REQUEST  0xc0000009

//
// MessageId: MSG_RECEIVED_STOP_COMMAND
//
#define MSG_RECEIVED_STOP_COMMAND  0x4000000a

//
// MessageId: MSG_AGENT_START_FAILURE
//
#define MSG_AGENT_START_FAILURE  0xc000000b

//
// MessageId: MSG_RECEIVED_STOP_SVC_COMMAND
//
#define MSG_RECEIVED_STOP_SVC_COMMAND  0x4000000c

//
// MessageId: MSG_RECEIVED_STOP_SHUTDOWN
//
#define MSG_RECEIVED_STOP_SHUTDOWN  0x4000000d

//
// MessageId: MSG_AGENT_SHUTDOWN_STARTING
//
#define MSG_AGENT_SHUTDOWN_STARTING  0x4000000e

//
// MessageId: MSG_WARNING_PROGRAMDATA_ERROR
//
#define MSG_WARNING_PROGRAMDATA_ERROR  0x8000000f

//
// MessageId: MSG_AGENT_PRE_SHUTDOWN_STARTING
//
#define MSG_AGENT_PRE_SHUTDOWN_STARTING  0x40000010

//
// MessageId: MSG_SYSPROBE_RESTART_INACTIVITY
//
#define MSG_SYSPROBE_RESTART_INACTIVITY  0x80000011

