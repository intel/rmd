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

/**
 * Maintains a table of core and class_id that are selected in config string for
 * setting up allocation policy per core
 */
typedef struct
{
    unsigned core;
    unsigned class_id;
} assoc_core;

/**
 * Maintains a table of core and pid that are selected in config string for
 * setting up allocation policy per COS
 */
typedef struct
{
    pid_t pid;
    unsigned class_id;
} assoc_pid;

/*
  The main structure needed to be filled by RMD to allocate CAT/MBA data
*/
typedef struct
{
    assoc_pid pid[PQOS_MAX_PID];
    assoc_core core[PQOS_MAX_CORES];
    unsigned int num_of_pid;
    unsigned int num_of_cores;

    struct pqos_l3ca sel_l3ca_cos[2]; // TODOThis must be based on number of sockets. For now hardcoding to 2. But this should be modified.

    // struct pqos_mba sel_mba_cos[];
} pqos_wrapper;

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

/*Associate core*/
int pqos_wrapper_assoc_core(assoc_core *core, unsigned int num_of_core)
{
    unsigned int i;

    for (i = 0; i < num_of_core; i++)
    {

        int ret;
        ret = pqos_alloc_assoc_set(core[i].core,
                                   core[i].class_id);
        if (ret != PQOS_RETVAL_OK)
        {
            debug_print("assoc_core failed!\n");
            return PQOS_RETVAL_ERROR;
        }
    }
    return PQOS_RETVAL_OK;
}

/*Asscoiate pid or task*/
int pqos_wrapper_assoc_pid(assoc_pid *pid, unsigned int num_of_pid)
{
    unsigned int i;
    for (i = 0; i < num_of_pid; i++)
    {

        int ret;
        ret = pqos_alloc_assoc_set_pid(pid[i].pid,
                                       pid[i].class_id);
        if (ret != PQOS_RETVAL_OK)
        {
            debug_print("assoc_pid class of service failed\n");
            return PQOS_RETVAL_ERROR;
        }
    }
    return PQOS_RETVAL_OK;
}

/*Allocate L3Cache*/
int pqos_wrapper_alloc_l3cache(struct pqos_l3ca *l3ca)
{
    int ret;
    const struct pqos_cpuinfo *p_cpu = NULL;
    const struct pqos_cap *p_cap = NULL;
    unsigned l3cat_id_count, *p_l3cat_ids = NULL;

    /* Get CMT capability and CPU info pointer */
    ret = pqos_cap_get(&p_cap, &p_cpu);
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

    for (unsigned int i = 0; i < l3cat_id_count; i++)
    {
        ret = pqos_l3ca_set(*p_l3cat_ids,
                            1,
                            &l3ca[i]);
        if (ret != PQOS_RETVAL_OK)
        {
            debug_print("Setting up cache allocation class of "
                        "service failed!\n");
            return PQOS_RETVAL_ERROR;
        }
        p_l3cat_ids++;
    }

    return PQOS_RETVAL_OK;
}

// pqos_wrapper *alloc_rdt
int pqos_wrapper_main(int clos, int pid, int core, int s0, int s1, int num_of_pid, int num_of_cores)
{
    debug_print("clos : %d\n", clos);
    // puts(clos);
    // fflush(stdout);
    int ret;
    struct pqos_config cfg;
    // memset(&cfg, 0, sizeof(cfg));
    // cfg.fd_log = STDOUT_FILENO;
    // cfg.verbose = 0;
    // cfg.interface = PQOS_INTER_OS;
    // ret = pqos_wrapper_init(&cfg);
    // if(ret != PQOS_RETVAL_OK)
    // {
    //     printf("Error initializing PQoS library!\n");
    //     	/* reset and deallocate all the resources */
    // }
    // else
    // {
    //     printf("pqos init successful\n");
    // }

    pqos_wrapper *alloc_rdt = (pqos_wrapper *)malloc(sizeof(pqos_wrapper));
    alloc_rdt->pid[0].pid = pid;
    alloc_rdt->pid[0].class_id = clos;
    alloc_rdt->num_of_pid = num_of_pid;

    /*Socket 0 cache allocation*/
    alloc_rdt->sel_l3ca_cos[0].class_id = clos;
    alloc_rdt->sel_l3ca_cos[0].u.ways_mask = s0;
    /*Socket 1 cache allocation*/
    alloc_rdt->sel_l3ca_cos[1].class_id = clos;
    alloc_rdt->sel_l3ca_cos[1].u.ways_mask = s1;

    alloc_rdt->core[0].core = 2;
    alloc_rdt->core[0].class_id = clos;
    // alloc_rdt.core[1].core=3;
    // alloc_rdt.core[1].class_id = 1;
    // alloc_rdt.num_of_cores=2;
    ret = pqos_wrapper_alloc_l3cache(alloc_rdt->sel_l3ca_cos);
    if (ret != PQOS_RETVAL_OK)
    {
        debug_print("Cache allocation failure!\n");
        /* reset and deallocate all the resources */
    }
    else
    {
        debug_print("Cache allocation success\n");
    }

    // ret = pqos_wrapper_assoc_core(alloc_rdt->core,alloc_rdt->num_of_cores);
    // if(ret != PQOS_RETVAL_OK)
    // {
    //     printf("Core association failure!\n");
    //     	/* reset and deallocate all the resources */
    // }
    // else
    // {
    //     printf("Core association  success\n");
    // }
    if (num_of_pid > 0)
    {
        ret = pqos_wrapper_assoc_pid(alloc_rdt->pid, alloc_rdt->num_of_pid);
        if (ret != PQOS_RETVAL_OK)
        {
            debug_print("pid association failure!\n");
            /* reset and deallocate all the resources */
        }
        else
        {
            debug_print("pid association  success\n");
        }
    }

    return ret;
}
