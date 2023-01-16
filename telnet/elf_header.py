import base64
import random

archs = {
    "EM_NONE":         0,
    "EM_M32":          1,
    "EM_SPARC":        2,
    "EM_386":          3,
    "EM_68K":          4, # m68k
    "EM_88K":          5, # m68k
    "EM_486":          6, # x86
    "EM_860":          7, # Unknown
    "EM_MIPS":         8,       # MIPS R3000 (officially, big-endian only) 
                                    # Next two are historical and binaries and
                                    #modules of these types will be rejected by
                                    #Linux.  
    "EM_MIPS_RS3_LE":  10,      # MIPS R3000 little-endian 
    "EM_MIPS_RS4_BE":  10,      # MIPS R4000 big-endian 

    "EM_PARISC":       15,      # HPPA 
    "EM_SPARC32PLUS":  18,      # Sun's "v8plus" 
    "EM_PPC":          20,      # PowerPC 
    "EM_PPC64":        21,      # PowerPC64 
    "EM_SPU":          23,      # Cell BE SPU 
    "EM_ARM":          40,      # ARM 32 bit 
    "EM_SH":           42,      # SuperH 
    "EM_SPARCV9":      43,      # SPARC v9 64-bit 
    "EM_H8_300":       46,      # Renesas H8/300 
    "EM_IA_64":        50,      # HP/Intel IA-64 
    "EM_X86_64":       62,      # AMD x86-64 
    "EM_S390":         22,      # IBM S/390 
    "EM_CRIS":         76,      # Axis Communications 32-bit embedded processor 
    "EM_M32R":         88,      # Renesas M32R 
    "EM_MN10300":      89,      # Panasonic/MEI MN10300, AM33 
    "EM_OPENRISC":     92,     # OpenRISC 32-bit embedded processor 
    "EM_BLACKFIN":     106,     # ADI Blackfin Processor 
    "EM_ALTERA_NIOS2": 113,     # Altera Nios II soft-core processor 
    "EM_TI_C6000":     140,     # TI C6X DSPs 
    "EM_AARCH64":      183,    # ARM 64 bit 
    "EM_TILEPRO":      188,     # Tilera TILEPro 
    "EM_MICROBLAZE":   189,     # Xilinx MicroBlaze 
    "EM_TILEGX":       191,     # Tilera TILE-Gx 
    "EM_FRV":          0x5441,  # Fujitsu FR-V 
    "EM_AVR32":        0x18ad,  # Atmel AVR32 
}

endian = {
    "EE_LITTLE":   1, # Little endian
    "EE_BIG":      2 # Big endian
}

types = {
    "ET_NOFILE":   0, # None
    "ET_REL":      1, # Relocatable file
    "ET_EXEC":     2, # Executable file
    "ET_DYN":      3, # Shared object file
    "ET_CORE":     4 # Core file
}

def random_elf_header(elf_template):
    endianness, endian_byte = random.choice(list(endian.items()))
    architecture, arch_code = random.choice(list(archs.items()))
    elf = bytearray(elf_template)
    elf[5] = endian_byte.to_bytes(1, 'little')[0]
    arch_bytes = arch_code.to_bytes(2, 'little')
    elf[18] = arch_bytes[0]
    elf[19] = arch_bytes[1]
    return (bytes(elf), endianness, architecture)
