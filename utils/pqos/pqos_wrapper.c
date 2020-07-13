#include "_cgo_export.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <ctype.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include "pqos.h"

// Easy switching between final code and additional debug prints
//#define DEBUG_PRINT
#ifdef DEBUG_PRINT
#define debug_print(...) printf(__VA_ARGS__)
#else
#define debug_print(...)
#endif

/**
 * Defines
 */
/**
 * Defines
 */
#define PQOS_MAX_CORES (1024)
#define PQOS_MAX_PID (128)
/*We need 4 main parameters for CAT & MBA. 
 * 1. CPU core information
 * 2. Task ID Information
 * 3. CDM for a particular COS
 * 4. MBA value for a particular COS*/

int pqos_wrapper_init();
int pqos_wrapper_check_mba_support(int *mbaMode);
int pqos_wrapper_finish();
int pqos_wrapper_reset_api();
int pqos_wrapper_alloc_release(const unsigned *core_array, unsigned int core_amount_in_array);
int pqos_wrapper_alloc_assign(const unsigned *core_array, unsigned int core_amount_in_array, unsigned *class_id);
int pqos_wrapper_set_mba_for_common_cos(unsigned classID, int mbaMode, const unsigned *mbaMax, const unsigned *socketsToSetArray, int numOfSockets);
int pqos_wrapper_alloc_l3cache(unsigned classID, const unsigned *waysMask, const unsigned *socketsToSet, int numOfSockets);
int pqos_wrapper_assoc_core(const unsigned *classIDs, const unsigned *cores, int numOfCores);
int pqos_wrapper_assoc_pid(const unsigned *classIDs, const unsigned *tasks, int numOfTasks);

// MBA struct type needed to set MBA correctly
// => REQUESTED - defines what mba value should be applied
// => ACTUAL - will be set by the PQoS library
enum mba_type
{
    REQUESTED = 0,
    ACTUAL,
    MAX_MBA_TYPES
};

/*PQOS init*/
int pqos_wrapper_init()
{
    int ret;
    struct pqos_config cfg;
    memset(&cfg, 0, sizeof(cfg));
    cfg.fd_log = STDOUT_FILENO;
    cfg.verbose = 0;
    cfg.interface = PQOS_INTER_OS;
    ret = pqos_init(&cfg);
    if (ret != PQOS_RETVAL_OK)
    {
        debug_print("Error initializing PQoS library!\n");
        ret = pqos_fini();
        if (ret != PQOS_RETVAL_OK)
            debug_print("Error shutting down PQoS library!\n");
        return PQOS_RETVAL_ERROR;
    }
    return PQOS_RETVAL_OK;
}

/* Associate core*/
// Example: params ({1,2,2}, {5,6,7}, 3) will be interpreted as following 3 pairs of (classid,coreid): (1,5),(2,6),(2,7)
int pqos_wrapper_assoc_core(const unsigned *classIDs, const unsigned *cores, int numOfCores)
{
    for (int i = 0; i < numOfCores; i++)
    {
        int ret;
        ret = pqos_alloc_assoc_set(cores[i],classIDs[i]);
        if (ret != PQOS_RETVAL_OK)
        {
            debug_print("assoc_core failed!\n");
            return PQOS_RETVAL_ERROR;
        }
    }
    return PQOS_RETVAL_OK;
}

/*Asscoiate pid or task*/
int pqos_wrapper_assoc_pid(const unsigned *classIDs, const unsigned *tasks, int numOfTasks)
{
    for (int i = 0; i < numOfTasks; i++)
    {
        int ret = pqos_alloc_assoc_set_pid(tasks[i], classIDs[i]);

        if (ret != PQOS_RETVAL_OK)
        {
            debug_print("assoc_pid class of service failed\n");
            return PQOS_RETVAL_ERROR;
        }
    }
    return PQOS_RETVAL_OK;
}

/*Allocate L3Cache*/
int pqos_wrapper_alloc_l3cache(unsigned classID, const unsigned *waysMask,const unsigned *socketsToSet, int numOfSockets)
{
    const struct pqos_cpuinfo *p_cpu = NULL;
    const struct pqos_cap *p_cap = NULL;
    unsigned l3cat_id_count, *p_l3cat_ids = NULL;

    /* Get CMT capability and CPU info pointer */
    int ret = pqos_cap_get(&p_cap, &p_cpu);
    if (ret != PQOS_RETVAL_OK)
    {
        debug_print("Error retrieving PQoS capabilities!\n");
        return PQOS_RETVAL_ERROR;
    }

    /* Get CPU l3cat id information to set COS */
    p_l3cat_ids = pqos_cpu_get_l3cat_ids(p_cpu, &l3cat_id_count);
    if (p_l3cat_ids == NULL)
    {
        debug_print("Error retrieving CPU socket information!\n");
        return PQOS_RETVAL_ERROR;
    }

    struct pqos_l3ca l3ca;
    for (int i = 0; i < numOfSockets; i++)
    {
        l3ca.class_id = classID;
        l3ca.u.ways_mask = (uint64_t)waysMask[i];

        int socket = socketsToSet[i];

        ret = pqos_l3ca_set(p_l3cat_ids[socket], 1, &l3ca);
        if (ret != PQOS_RETVAL_OK)
        {
            debug_print("Setting up cache allocation class of service failed!\n");
            return PQOS_RETVAL_ERROR;
        }
    }

    return PQOS_RETVAL_OK;
}

// Checks if MBA is supported and in which mode
// [out] mbaMode where 0 percentage mode
//                     1 MBps mode
//                    -1 MBA is not supported
// [out] PQOS_RETVAL_OK when no error
//       PQOS_RETVAL_ERROR when errorr
int pqos_check_mba_support(int *mbaMode)
{
    const struct pqos_cap *p_cap = NULL;
    const struct pqos_cpuinfo *p_cpu = NULL;
    const struct pqos_capability *cap_mba = NULL;

    /* Get CMT capability and CPU info pointer */
    int ret = pqos_cap_get(&p_cap, &p_cpu);
    if (ret != PQOS_RETVAL_OK)
    {
        printf("Error retrieving PQoS capabilities!\n");
        return ret;
    }

    ret = pqos_cap_get_type(p_cap, PQOS_CAP_TYPE_MBA, &cap_mba);

    if (ret != PQOS_RETVAL_OK)
    {
        debug_print("Failed to get pqos_capability struct for MBA\n");
        return PQOS_RETVAL_ERROR;
    }

    if (cap_mba == NULL)
    {
        debug_print("MBA not supported\n");
        *mbaMode = -1;
        return PQOS_RETVAL_OK;
    }

    debug_print("MBA is supported\n");

    if (cap_mba->u.mba != NULL)
    {
        if (cap_mba->u.mba->ctrl == 0)
        {
            debug_print("MBA in percentage mode\n");
            *mbaMode = 0;
        }
        else if (cap_mba->u.mba->ctrl == 1)
        {
            debug_print("MBA in MBps mode\n");
            *mbaMode = 1;
        }
        else
        {
            debug_print("MBA in not in MBps or percentage mode\n");
            return PQOS_RETVAL_ERROR;
        }
    }
    return PQOS_RETVAL_OK;
}

/*Shuts down PQoS module*/
int pqos_wrapper_finish()
{
    return pqos_fini();
}

/*Reset PQoS API*/
int pqos_wrapper_reset_api()
{
    return pqos_alloc_reset(PQOS_REQUIRE_CDP_ANY, PQOS_REQUIRE_CDP_ANY, PQOS_MBA_ANY);
}

/*Reassign cores in core_array to default COS#0 - please be aware that function
  will not reset COS params to default values because releasing core from COS is enough
   @param  [in] core_array            list of core ids
   @param  [in] core_amount_in_array  number of core ids in the core_array
*/
int pqos_wrapper_alloc_release(const unsigned *core_array, unsigned int core_amount_in_array)
{
    return pqos_alloc_release(core_array, core_amount_in_array);
}

/*
   Assign first available COS to cores in core_array
   @param [in]  core_array   list of core ids
   @param [in]  core_num     number of core ids in the core_array
   @param [out] class_id     place to store reserved COS id
   @return operation status
*/
int pqos_wrapper_alloc_assign(const unsigned *core_array, unsigned int core_amount_in_array, unsigned int *class_id)
{
    return pqos_alloc_assign((1 << PQOS_CAP_TYPE_L3CA | 1 << PQOS_CAP_TYPE_MBA), core_array, core_amount_in_array, class_id);
}

/*Sets classes of service defined by mba on mba id for common COS#
    @param [in]  classID             class of service
    @param [in]  mbaMode             common mbaMode for all sockets: 0 for percentage mode or 1 for MBps mode
    @param [in]  mbaMax              mba values to set for all sockets (ascending socket number order)
    @param [in]  socketsToSetArray   sockets to set for common COS# (ascending socket number order)
    @param [in]  numOfSockets        amount of values in mbMaxesArray which should be also amount of elements in socketsToSetArray
    @return operation status - PQOS_RETVAL_OK on success
*/
int pqos_wrapper_set_mba_for_common_cos(unsigned classID, int mbaMode, const unsigned *mbaMax, const unsigned *socketsToSetArray, int numOfSockets)
{
    const struct pqos_cpuinfo *p_cpu = NULL;
    const struct pqos_cap *p_cap = NULL;
    unsigned mba_id_count, *p_mba_ids = NULL;

    /* Get CMT capability and CPU info pointer */
    int ret = pqos_cap_get(&p_cap, &p_cpu);
    if (ret != PQOS_RETVAL_OK)
    {
        debug_print("Error retrieving PQoS capabilities!\n");
        return PQOS_RETVAL_ERROR;
    }
    /* Get CPU mba_id information to set COS */
    p_mba_ids = pqos_cpu_get_mba_ids(p_cpu, &mba_id_count);
    if (p_mba_ids == NULL)
    {
        debug_print("Error retrieving MBA ID information!\n");
        return PQOS_RETVAL_ERROR;
    }

    // Table containing  MBA "requested" and "actual" COS definitions represented here by maxMBATypes
    // => "requested" - defines values to set (mb_max)
    // => "actual" - set by the PQoS library
    struct pqos_mba mba[MAX_MBA_TYPES];
    for (int i = 0; i < numOfSockets; i++)
    {
        mba[REQUESTED].class_id = classID;
        mba[REQUESTED].mb_max = mbaMax[i];
        mba[REQUESTED].s = mbaMode;

        int socket = socketsToSetArray[i];
        ret = pqos_mba_set(p_mba_ids[socket], 1, &mba[REQUESTED], &mba[ACTUAL]);

        if (ret != PQOS_RETVAL_OK)
        {
            debug_print("Failed to set MBA!\n");
            return PQOS_RETVAL_ERROR;
        }
    }

    debug_print("MBA allocation success\n");
    return PQOS_RETVAL_OK;
}