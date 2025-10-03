
/* RescueTime::Network::API::activate_enterprise(std::__cxx11::basic_string<wchar_t,
   std::char_traits<wchar_t>, std::allocator<wchar_t> > const&) */

API * __thiscall RescueTime::Network::API::activate_enterprise(API *this,basic_string *param_1)

{
  size_t sVar1;
  long in_FS_OFFSET;
  undefined *local_e8 [2];
  undefined local_d8 [16];
  undefined *local_c8 [2];
  undefined local_b8 [16];
  undefined *local_a8 [2];
  undefined auStack_98 [16];
  wchar_t *local_88;
  int local_80;
  wchar_t local_78 [4];
  undefined *local_68;
  undefined local_58 [24];
  long local_40;

  local_40 = *(long *)(in_FS_OFFSET + 0x28);
  local_88 = local_78;
  sVar1 = wcslen(L"[API::activate_enterprise]");
  std::__cxx11::basic_string<>::_M_construct<>
            ((wchar_t *)&local_88,L"[API::activate_enterprise]",(int)sVar1 * 4 + 0xa1e750);
                    /* try { // try from 00631131 to 00631135 has its CatchHandler @ 00631332 */
  rt::debug_log((basic_string *)&local_88,0xc0);
  if (local_88 != local_78) {
    operator.delete(local_88);
  }
  local_a8[0] = auStack_98;
  sVar1 = wcslen(L"");
  std::__cxx11::basic_string<>::_M_construct<>((wchar_t *)local_a8,L"",(int)sVar1 * 4 + 0x9f06ec);
  local_c8[0] = local_b8;
  sVar1 = wcslen(L"");
                    /* try { // try from 006311b5 to 006311b9 has its CatchHandler @ 0063132d */
  std::__cxx11::basic_string<>::_M_construct<>((wchar_t *)local_c8,L"",(int)sVar1 * 4 + 0x9f06ec);
  local_e8[0] = local_d8;
  sVar1 = wcslen(L"");
                    /* try { // try from 006311e7 to 006311eb has its CatchHandler @ 00631328 */
  std::__cxx11::basic_string<>::_M_construct<>((wchar_t *)local_e8,L"",(int)sVar1 * 4 + 0x9f06ec);
                    /* try { // try from 0063120c to 00631210 has its CatchHandler @ 006312d7 */
  perform_activate((basic_string *)&local_88,param_1,(basic_string *)local_e8,
                   (basic_string *)local_c8);
  if (local_e8[0] != local_d8) {
    operator.delete(local_e8[0]);
  }
  if (local_c8[0] != local_b8) {
    operator.delete(local_c8[0]);
  }
  if (local_a8[0] != auStack_98) {
    operator.delete(local_a8[0]);
  }
  *(API **)this = this + 0x10;
                    /* try { // try from 00631273 to 00631277 has its CatchHandler @ 006312c4 */
  std::__cxx11::basic_string<>::_M_construct<>
            ((wchar_t *)this,local_88,(int)local_88 + local_80 * 4);
  if (local_68 != local_58) {
    operator.delete(local_68);
  }
  if (local_88 != local_78) {
    operator.delete(local_88);
  }
  if (local_40 == *(long *)(in_FS_OFFSET + 0x28)) {
    return this;
  }
                    /* WARNING: Subroutine does not return */
  __stack_chk_fail();
}

